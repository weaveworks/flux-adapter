package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
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
	fs := pflag.NewFlagSet("default", pflag.ContinueOnError)

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
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		fs.Usage()
		os.Exit(1)
	}
}
