package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/s5i/goutil/shutdown"
	"github.com/s5i/goutil/version"
	"github.com/s5i/ruuvi2db/licenses"
	"github.com/s5i/ruuvi2db/reader"
	"golang.org/x/sync/errgroup"

	_ "net/http/pprof"
)

var (
	fConfig      = flag.String("config", "", "Path to config file.")
	fLicenses    = flag.Bool("licenses", false, "When true, print attached licenses and exit.")
	fVersion     = flag.Bool("version", false, "When true, print version and exit.")
	fDebugListen = flag.String("debug_listen", "", "If set, opens a debug port.")
)

func main() {
	flag.Parse()

	if *fVersion {
		fmt.Fprintln(os.Stderr, version.Get())
		os.Exit(0)
	}

	if *fLicenses {
		fmt.Fprintln(os.Stderr, licenses.Merged())
		os.Exit(0)
	}

	if debugListen := *fDebugListen; debugListen != "" {
		// Not subject to context-based lifetime management in case we need to debug that ;)
		go func() {
			if err := http.ListenAndServe(debugListen, nil); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}()
	}

	ctx, cancel := context.WithCancel(context.Background())
	go shutdown.OnSignal(os.Interrupt, cancel)

	cfg, err := reader.ReadConfig(*fConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := cfg.Sanitize(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	g, ctx := errgroup.WithContext(ctx)
	defer func() {
		if err := g.Wait(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}()

	reader.Run(ctx, g, cfg)
}
