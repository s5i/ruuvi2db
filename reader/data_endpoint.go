package reader

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/s5i/ruuvi2db/data"
)

type RunDataEndpointOpts struct {
	Listen  string
	PointsF func() []*data.Point
}

func RunDataEndpoint(ctx context.Context, opts *RunDataEndpointOpts) error {
	srv := http.Server{}
	srv.Addr = opts.Listen

	srv.ReadTimeout = time.Minute
	srv.WriteTimeout = time.Minute
	srv.SetKeepAlivesEnabled(false)

	mux := http.NewServeMux()

	mux.Handle("/data.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		e := json.NewEncoder(w)
		e.SetIndent("", "  ")

		d := opts.PointsF()
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
