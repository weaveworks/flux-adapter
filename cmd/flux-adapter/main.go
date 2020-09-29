package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	_ "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/spf13/pflag"

	//	api "github.com/weaveworks/flux/api"
	transport "github.com/fluxcd/flux/pkg/http"
	fluxclient "github.com/fluxcd/flux/pkg/http/client"
	fluxserver "github.com/fluxcd/flux/pkg/http/daemon"
	"github.com/fluxcd/flux/pkg/remote"
)

type adapterConfig struct {
	// URL (either ws[s] or http[s]) for Weave Cloud
	connectURL  string
	apiURL      string
	bearerToken string
	listen      string
}

var version = "unversioned"

func main() {
	var conf adapterConfig
	fs := pflag.NewFlagSet(filepath.Base(os.Args[0]), pflag.ContinueOnError)

	versionFlag := fs.Bool("version", false, "print the version and exit")

	fs.StringVar(&conf.connectURL, "connect", "https://cloud.weave.works/api/flux", "Connect to Weave Cloud at this base address, including the path /api/flux")
	fs.StringVar(&conf.apiURL, "api", "http://localhost:3030/api/flux", "Connect to the flux API (i.e., fluxd) at this base URL, including if necessary the path /api/flux")
	fs.StringVar(&conf.bearerToken, "token", "", "use this bearer token to authenticate with Weave Cloud")
	fs.StringVar(&conf.listen, "listen", "localhost:3039", "address on which to listen for fluxd connections")

	err := fs.Parse(os.Args[1:])

	if *versionFlag {
		println(version)
		os.Exit(0)
	}

	if err != nil {
		if err == pflag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	relay := fluxclient.New(http.DefaultClient, transport.NewAPIRouter(), conf.apiURL, fluxclient.Token(""))
	up, err := fluxserver.NewUpstream(
		&http.Client{Timeout: 10 * time.Second},
		fmt.Sprintf("flux-adapter", version),
		fluxclient.Token(conf.bearerToken),
		transport.NewUpstreamRouter(),
		conf.connectURL,
		remote.NewErrorLoggingServer(relay, logger),
		10*time.Second,
		logger)

	if err != nil {
		logger.Log("err", err)
		os.Exit(1)
	}

	router := mux.NewRouter()
	transport.UpstreamRoutes(router)
	shutdown := make(chan struct{})
	handler, err := installHandlers(router, conf, logger, shutdown)
	if err != nil {
		logger.Log("err", err)
		os.Exit(1)
	}

	errc := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		m := http.DefaultServeMux
		m.Handle("/", handler)
		errc <- http.ListenAndServe(conf.listen, m)
	}()

	logger.Log("msg", "shutting down", "err", <-errc)
	close(shutdown)
	up.Close()
}

// This handler will serve the endpoint for events, and send them on
// to Weave Cloud. It includes the daemon RPC (websocket connection)
// endpoints for now, so that the fluxd argument `--connect` can be
// used to target this adapter. This will be removed eventually.

func installHandlers(r *mux.Router, config adapterConfig, logger log.Logger, shutdown chan struct{}) (*mux.Router, error) {
	serviceURL, err := url.Parse(config.connectURL)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(serviceURL)
	director := proxy.Director
	proxy.Director = func(r *http.Request) {
		director(r)
		fluxclient.Token(config.bearerToken).Set(r)
	}

	r.Get(transport.LogEvent).HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		logger.Log("info", "proxying", "url", req.URL.String())
		proxy.ServeHTTP(res, req)
	})

	upgrader := websocket.Upgrader{}

	for _, endpoint := range []string{
		transport.RegisterDaemonV6,
		transport.RegisterDaemonV7,
		transport.RegisterDaemonV8,
		transport.RegisterDaemonV9,
		transport.RegisterDaemonV10,
		transport.RegisterDaemonV11,
	} {
		r.Get(endpoint).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Log("msg", "websocket connection", "addr", r.RemoteAddr)
			ws, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, err.Error())
				return
			}

			// it's necessary to pull frames through the websocket, so
			// that controls get processed by the connection. (A close
			// from the other side will trigger an error in
			// NextReader.)
			wsErr := make(chan error, 1)
			go func() {
				for {
					if _, _, err = ws.NextReader(); err != nil {
						wsErr <- err
						break
					}
				}
			}()

			select {
			case <-shutdown:
				logger.Log("msg", "closing websocket", "reason", "shutting down")
			case err = <-wsErr:
				logger.Log("msg", "closing websocket", "reason", "errored", "err", err.Error())
			}

			ws.Close()
			logger.Log("msg", "closed websocket", "addr", r.RemoteAddr)
		})
	}

	return r, nil
}
