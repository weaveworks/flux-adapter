package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/spf13/pflag"

	//	api "github.com/weaveworks/flux/api"
	apiV9 "github.com/weaveworks/flux/api/v9"
	transport "github.com/weaveworks/flux/http"
	fluxclient "github.com/weaveworks/flux/http/client"
	fluxserver "github.com/weaveworks/flux/http/daemon"
	"github.com/weaveworks/flux/remote"
)

type adapterConfig struct {
	// URL (either ws[s] or http[s]) for Weave Cloud
	connectURL  string
	apiURL      string
	bearerToken string
}

var version = "unversioned"

func main() {
	var conf adapterConfig
	fs := pflag.NewFlagSet(filepath.Base(os.Args[0]), pflag.ContinueOnError)

	versionFlag := fs.Bool("version", false, "print the version and exit")

	fs.StringVar(&conf.connectURL, "connect", "https://cloud.weave.works/api/flux", "Connect to Weave Cloud at this base address, including the path /api/flux")
	fs.StringVar(&conf.apiURL, "api", "http://localhost:3030/api/flux", "Connect to the flux API (i.e., fluxd) at this base URL, including if necessary the path /api/flux")
	fs.StringVar(&conf.bearerToken, "token", "", "use this bearer token to authenticate with Weave Cloud")

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

	apiClient := fluxclient.New(http.DefaultClient, transport.NewAPIRouter(), conf.apiURL, fluxclient.Token(""))
	relay := &upstreamRelay{Client: apiClient}
	up, err := fluxserver.NewUpstream(
		&http.Client{Timeout: 10 * time.Second},
		fmt.Sprintf("flux-adapter", version),
		fluxclient.Token(conf.bearerToken),
		transport.NewUpstreamRouter(),
		conf.connectURL,
		remote.NewErrorLoggingUpstreamServer(relay, logger),
		10*time.Second,
		logger)

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

	logger.Log("msg", "shutting down", "err", <-errc)
	up.Close()
}

// To create an "upstream" (a component that will connect to Weave
// Cloud and relay RPCs), you need an implementation of
// `[...]flux/api/UpstreamServer`. This is the flux API, as well as
// some methods for checking connectivity (Ping, Version) and relaying
// webhook notifications (NotifyChange).
//
// Since we're going to just proxy API calls, a flux client (as used
// by fluxctl) will serve as the api.Server implementation. For
// api.Upstream, we'll have to have alternate implementations, or (at
// worst) stub out methods until provision can be made in fluxd's HTTP
// interface.

type upstreamRelay struct {
	*fluxclient.Client
}

var notImplemented = errors.New("not implemented")

func (r *upstreamRelay) Ping(ctx context.Context) error {
	return notImplemented
}

func (r *upstreamRelay) Version(ctx context.Context) (string, error) {
	return "1.14.0", nil
}

func (r *upstreamRelay) NotifyChange(ctx context.Context, change apiV9.Change) error {
	return notImplemented
}
