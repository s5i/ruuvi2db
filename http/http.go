package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/s5i/ruuvi2db/config"
	"github.com/s5i/ruuvi2db/data"
)

func Run(ctx context.Context, cfg *config.Config, db DB) error {
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

	http.Handle("/", staticHandler())
	http.Handle("/data.json", dataHandler(db))
	http.Handle("/tags.json", tagsHandler())

	if cfg.Debug.HTTPHandlers {
		http.Handle("/threadz", threadzHandler())
	}

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

type DB interface {
	Get(startTime, endTime time.Time) []data.Point
}
