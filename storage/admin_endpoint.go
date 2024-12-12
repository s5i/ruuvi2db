package storage

import (
	"context"
	"errors"
	"net/http"
)

type RunAdminEndpointOpts struct {
	Listen    string
	SetAliasF func(addr, name string) error
}

func RunAdminEndpoint(ctx context.Context, opts *RunAdminEndpointOpts) error {
	srv := http.Server{}
	srv.Addr = opts.Listen

	mux := http.NewServeMux()
	mux.Handle("/admin/set_alias", SetAliasHandler(&SetAliasHandlerOpts{
		SetAliasF: opts.SetAliasF,
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

type SetAliasHandlerOpts struct {
	SetAliasF func(addr, name string) error
}

func SetAliasHandler(opts *SetAliasHandlerOpts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		addr, ok, err := singleStringParam(r, "addr")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if !ok {
			http.Error(w, "addr not specified", 500)
		}

		name, _, err := singleStringParam(r, "name")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if err := opts.SetAliasF(addr, name); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
}
