package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/s5i/goutil/shutdown"
	"github.com/s5i/goutil/version"
	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/licenses"
	"github.com/s5i/ruuvi2db/storage/database/bolt"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"

	_ "net/http/pprof"
)

var (
	fConfig         = flag.String("config", "", "Path to config file.")
	fLicenses       = flag.Bool("licenses", false, "When true, print attached licenses and exit.")
	fVersion        = flag.Bool("version", false, "When true, print version and exit.")
	fAllowRewriteDB = flag.Bool("allow_rewrite_db", false, "Allows database to be rewritten to a newer schema.")
)

type config struct {
	HTTP struct {
		Listen      string `yaml:"listen"`
		DebugListen string `yaml:"debug_listen"`
		AdminListen string `yaml:"admin_listen"`
	} `yaml:"http"`

	DataSource struct {
		Address      string        `yaml:"address"`
		QueryPeriod  time.Duration `yaml:"query_period"`
		MaxStaleness time.Duration `yaml:"max_staleness"`
		MACFilter    []string      `yaml:"mac_filter"`
	} `yaml:"data_source"`

	Storage struct {
		Bolt struct {
			Path            string        `yaml:"path"`
			RetentionWindow time.Duration `yaml:"retention_window"`
		} `yaml:"bolt"`
	} `yaml:"storage"`
}

func main() {
	flag.Parse()

	if *fLicenses {
		fmt.Fprintln(os.Stderr, licenses.Merged())
		os.Exit(0)
	}

	if *fVersion {
		fmt.Fprintln(os.Stderr, version.Get())
		return
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

	db := bolt.New()

	g, ctx := errgroup.WithContext(ctx)
	defer func() {
		if err := g.Wait(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}()

	g.Go(func() error {
		return db.Run(ctx, &bolt.Config{
			Path:            cfg.Storage.Bolt.Path,
			RetentionWindow: cfg.Storage.Bolt.RetentionWindow,
			AllowRewrite:    *fAllowRewriteDB,
		})
	})

	g.Go(func() error {
		return runSourcePoller(ctx, cfg.DataSource.Address, cfg.DataSource.QueryPeriod, cfg.DataSource.MaxStaleness, cfg.DataSource.MACFilter, db.PushPoints)
	})

	g.Go(func() error {
		return runHTTP(ctx, cfg.HTTP.Listen, db.Points, db.Alias)
	})

	if cfg.HTTP.AdminListen != "" {
		g.Go(func() error {
			return runAdminHTTP(ctx, cfg.HTTP.AdminListen, db.SetAlias)
		})
	}
}

func readConfig(path string) (*config, error) {
	switch {
	case path != "":
		path = sanitizePath(path)
	case os.Geteuid() == 0:
		path = "/usr/local/ruuvi2db/storage.cfg"
	default:
		path = fmt.Sprintf("%s/.ruuvi2db/storage.cfg", os.Getenv("HOME"))
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

	cfg.Storage.Bolt.Path = sanitizePath(cfg.Storage.Bolt.Path)

	return cfg, nil
}

func sanitizePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(os.Getenv("HOME"), strings.TrimPrefix(path, "~/"))
	}
	return path
}

func runSourcePoller(ctx context.Context, addr string, period time.Duration, maxStaleness time.Duration, macFilter []string, pushPoints func([]*data.Point) error) error {
	endpoint, err := url.JoinPath("http://", addr, "data.json")
	if err != nil {
		return err
	}

	filter := map[string]bool{}
	for _, mac := range macFilter {
		filter[strings.ToUpper(mac)] = true
	}

	tick := time.NewTicker(period)
	for {
		func() {
			resp, err := http.Get(endpoint)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			var src, dst []*data.Point
			d := json.NewDecoder(resp.Body)
			if err := d.Decode(&src); err != nil {
				return
			}

			for _, p := range src {
				if len(macFilter) > 0 && !filter[strings.ToUpper(p.Address)] {
					continue
				}

				p.Timestamp = p.Timestamp.Truncate(period)
				if p.Timestamp.Add(maxStaleness).Before(time.Now()) {
					continue
				}

				dst = append(dst, p)
			}

			if err := pushPoints(dst); err != nil {
				log.Print(err)
				return
			}
		}()

		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
		}
	}
}

func runHTTP(ctx context.Context, listen string, pointsBetween func(startTime, endTime time.Time) ([]*data.Point, error), addrToName func(string) (string, error)) error {
	srv := http.Server{}
	srv.Addr = listen

	srv.ReadTimeout = time.Minute
	srv.WriteTimeout = time.Minute
	srv.SetKeepAlivesEnabled(false)

	mux := http.NewServeMux()

	mux.Handle("/data.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endTime, err := dataEndTime(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		duration, err := dataDuration(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		kind, err := dataKind(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		src, err := pointsBetween(endTime.Add(-duration), endTime)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if endTime.Before(time.Now()) {
			w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
		}

		m := map[time.Time]map[string]any{}
		for _, p := range src {
			if m[p.Timestamp] == nil {
				m[p.Timestamp] = map[string]any{}
			}

			name, err := addrToName(p.Address)
			if err != nil || name == "" {
				name = p.Address
			}

			m[p.Timestamp]["ts"] = p.Timestamp
			m[p.Timestamp][name] = dataValue(p, kind)
		}

		ret := []map[string]any{}
		for _, v := range m {
			ret = append(ret, v)
		}

		sort.Slice(ret, func(i, j int) bool { return ret[i]["ts"].(time.Time).Before(ret[j]["ts"].(time.Time)) })

		e := json.NewEncoder(w)
		e.SetIndent("", "  ")

		if err := e.Encode(ret); err != nil {
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

func runAdminHTTP(ctx context.Context, listen string, setAlias func(addr, name string) error) error {
	srv := http.Server{}
	srv.Addr = listen

	mux := http.NewServeMux()
	mux.Handle("/admin/set_alias", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addr, err := singleStringParam(r, "addr")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		name, err := singleStringParam(r, "name")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if err := setAlias(addr, name); err != nil {
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

func dataValue(p *data.Point, kind string) any {
	switch kind {
	case "temperature":
		return fmt.Sprintf("%.2f", p.Temperature)
	case "humidity":
		return fmt.Sprintf("%.2f", p.Humidity)
	case "pressure":
		return fmt.Sprintf("%.2f", p.Pressure)
	case "battery":
		return fmt.Sprintf("%.2f", p.Battery)
	}
	return nil
}

func singleStringParam(r *http.Request, p string) (string, error) {
	x, ok := r.URL.Query()[p]
	if !ok {
		return "", fmt.Errorf("%q not specified", p)
	}

	if len(x) != 1 {
		return "", fmt.Errorf("%q specified multiple times", p)
	}

	return x[0], nil
}

func dataKind(r *http.Request) (string, error) {
	x, err := singleStringParam(r, "kind")
	if err != nil {
		return "", err
	}

	switch x {
	case "temperature", "humidity", "pressure", "battery":
	default:
		return "", fmt.Errorf("unrecognized kind %q", x)
	}

	return x, nil
}

func dataEndTime(r *http.Request) (time.Time, error) {
	x, err := singleStringParam(r, "end_time")
	if err != nil {
		return time.Time{}, err
	}

	ret, err := strconv.ParseInt(x, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("malformed end_time %q", x)
	}

	return time.Unix(ret, 0), nil
}

func dataDuration(r *http.Request) (time.Duration, error) {
	x, err := singleStringParam(r, "duration")
	if err != nil {
		return 0, err
	}

	ret, err := strconv.ParseInt(x, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("malformed duration %q", x)
	}

	return time.Duration(ret) * time.Second, nil
}
