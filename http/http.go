package http

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/s5i/ruuvi2db/config"
	"github.com/s5i/ruuvi2db/data"
)

//go:embed static/*
var StaticData embed.FS

func Run(ctx context.Context, cfg *config.Config, src DB) error {
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

	files, err := StaticData.ReadDir("static")
	if err != nil {
		return fmt.Errorf("failed to read embedded data: %v", err)
	}

	for _, f := range files {
		fName := f.Name()
		content, err := StaticData.ReadFile("static/" + fName)
		if err != nil {
			return fmt.Errorf("failed to read embedded data: %v", err)
		}
		if fName == "index.html" {
			fName = ""
		}

		http.HandleFunc("/"+fName, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "public, max-age=360, immutable")
			w.Header().Set("Content-Type", contentType(fName))
			if _, err := w.Write(content); err != nil {
				http.Error(w, "Something went wrong", 500)
			}
		})
	}

	http.Handle("/csv", newCSVHandler(src))
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
	Get(startTime, endTime time.Time) map[string][]data.Point
}
