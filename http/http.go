package http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/s5i/ruuvi2db/config"
	"github.com/s5i/ruuvi2db/storage/database/bolt"
)

func Run(ctx context.Context, cfg *config.Config, db *bolt.DB) error {
	if !cfg.HTTP.Enable {
		return nil
	}

	srv := http.Server{}

	listen := cfg.HTTP.Listen
	if listen == "" {
		if os.Geteuid() == 0 {
			listen = ":80"
		} else {
			listen = ":8080"
		}
	}
	srv.Addr = listen

	srv.ReadTimeout = time.Minute
	srv.WriteTimeout = time.Minute
	srv.SetKeepAlivesEnabled(false)

	mux := http.NewServeMux()

	mux.Handle("/", staticHandler())
	mux.Handle("/data.json", dataHandler(db))
	mux.Handle("/tags.json", tagsHandler())

	if cfg.Debug.HTTPHandlers {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	srv.Handler = mux

	go func() {
		<-ctx.Done()
		srv.Shutdown(ctx)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("srv.ListenAndServe: %v", err)
	}

	return nil
}

func contentType(fName string) string {
	split := strings.Split(fName, ".")
	ext := split[len(split)-1]

	switch ext {
	case "", "html":
		return "text/html"
	case "js":
		return "application/javascript"
	case "css":
		return "text/css"
	default:
		return "text/plain"
	}
}
