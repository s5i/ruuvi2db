package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/s5i/ruuvi2db/config"
	"github.com/s5i/ruuvi2db/db"
)

func Run(ctx context.Context, cfg *config.Config, dbs map[string]db.Interface) error {
	srv := http.Server{}

	listen := cfg.GetHttp().Listen
	if listen == "" {
		if os.Geteuid() == 0 {
			listen = ":80"
		} else {
			listen = ":8080"
		}
	}
	srv.Addr = listen

	srcDB := cfg.GetHttp().SourceDb
	if _, ok := dbs[srcDB]; !ok {
		return fmt.Errorf("source_db (%q) not enabled", srcDB)
	}

	src, ok := dbs[srcDB].(db.Source)
	if !ok {
		return fmt.Errorf("reading from source_db (%q) not supported", srcDB)
	}

	for fName, content := range StaticData {
		fName, content := fName, content
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
