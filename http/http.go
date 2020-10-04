package http

import (
	"context"
	"fmt"
	"net/http"
	"os"

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
