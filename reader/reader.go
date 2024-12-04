package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/s5i/goutil/shutdown"
	"github.com/s5i/goutil/version"
	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/licenses"
	"github.com/s5i/ruuvi2db/reader/bluetooth"
	"github.com/s5i/ruuvi2db/reader/protocol"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"

	_ "net/http/pprof"
)

var (
	fConfig   = flag.String("config", "", "Path to config file.")
	fLicenses = flag.Bool("licenses", false, "When true, print attached licenses and exit.")
	fVersion  = flag.Bool("version", false, "When true, print version and exit.")
)

type config struct {
	HTTP struct {
		Listen      string `yaml:"listen"`
		DebugListen string `yaml:"debug_listen"`
	} `yaml:"http"`

	Bluetooth struct {
		WatchdogTimeout time.Duration `yaml:"watchdog_timeout"`
	} `yaml:"bluetooth"`

	Data struct {
		MaxStaleness time.Duration `yaml:"max_staleness"`
		MACFilter    []string      `yaml:"mac_filter"`
	} `yaml:"data"`
}

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

	ctx, cancel := context.WithCancel(context.Background())
	go shutdown.OnSignal(os.Interrupt, cancel)

	cfg, err := readConfig(*fConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if cfg.HTTP.DebugListen != "" {
		// Not subject to context-based lifetime management in case we need to debug that ;)
		go func() {
			if err := http.ListenAndServe(cfg.HTTP.DebugListen, nil); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}()
	}

	g, ctx := errgroup.WithContext(ctx)
	defer func() {
		if err := g.Wait(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}()

	get, put := makeCache(cfg.Data.MaxStaleness, cfg.Data.MACFilter)

	g.Go(func() error {
		return runBluetooth(ctx, cfg.Bluetooth.WatchdogTimeout, put)
	})

	g.Go(func() error {
		return runHTTP(ctx, cfg.HTTP.Listen, get)
	})
}

func readConfig(path string) (*config, error) {
	switch {
	case path != "":
	case os.Geteuid() == 0:
		path = "/usr/local/ruuvi2db/reader.cfg"
	default:
		path = fmt.Sprintf("%s/.ruuvi2db/reader.cfg", os.Getenv("HOME"))
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %v", path, err)
	}

	cfg := &config{}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %q: %v", path, err)
	}

	return cfg, nil
}

func runBluetooth(ctx context.Context, timeout time.Duration, f func(*data.Point)) error {
	switch err := bluetooth.Run(ctx, func(addr string, mfID uint16, datagram []byte) {
		p, err := protocol.ParseDatagram(mfID, datagram, addr)
		if err != nil {
			return
		}
		f(p)
	}, timeout); {
	case errors.Is(err, context.Canceled):
		return nil
	case errors.Is(err, bluetooth.ErrInit):
		return fmt.Errorf("%v\nDid you set the capability?\n$ sudo setcap cap_net_raw,cap_net_admin=ep /path/to/reader", err)
	default:
		return err
	}
}

func runHTTP(ctx context.Context, listen string, f func() []*data.Point) error {
	srv := http.Server{}
	srv.Addr = listen

	srv.ReadTimeout = time.Minute
	srv.WriteTimeout = time.Minute
	srv.SetKeepAlivesEnabled(false)

	mux := http.NewServeMux()

	mux.Handle("/data.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		e := json.NewEncoder(w)
		e.SetIndent("", "  ")

		d := f()
		if len(d) == 0 {
			d = []*data.Point{}
		}

		if err := e.Encode(d); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}))

	srv.Handler = mux

	go func() {
		<-ctx.Done()
		srv.Shutdown(ctx)
	}()

	switch err := srv.ListenAndServe(); {
	case errors.Is(err, http.ErrServerClosed):
		return nil
	default:
		return err
	}
}

func makeCache(maxStaleness time.Duration, macFilter []string) (get func() []*data.Point, put func(*data.Point)) {
	filter := map[string]bool{}
	for _, m := range macFilter {
		filter[strings.ToUpper(m)] = true
	}

	var mu sync.Mutex
	points := map[string]*data.Point{}

	getF := func() []*data.Point {
		mu.Lock()
		defer mu.Unlock()

		var ret []*data.Point
		for k, v := range points {
			if v.Timestamp.Add(maxStaleness).Before(time.Now()) {
				delete(points, k)
				continue
			}
			ret = append(ret, v)
		}
		return ret
	}

	putF := func(p *data.Point) {
		mu.Lock()
		defer mu.Unlock()

		if len(filter) == 0 || filter[p.Address] {
			points[p.Address] = p
		}
	}
	return getF, putF
}
