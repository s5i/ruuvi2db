package ui

import (
	"context"
	"embed"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

type RunUIEndpointOpts struct {
	Listen      string
	StorageAddr string
}

func RunUIEndpoint(ctx context.Context, opts *RunUIEndpointOpts) error {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".json") {
			ProxyHandler(&ProxyHandlerOpts{
				StorageAddr: opts.StorageAddr,
			})(w, r)
			return
		}

		StaticHandler(&StaticHandlerOpts{})(w, r)
	}))

	srv := http.Server{
		Addr:         opts.Listen,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		Handler:      mux,
	}
	srv.SetKeepAlivesEnabled(false)

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

type ProxyHandlerOpts struct {
	StorageAddr string
}

func ProxyHandler(opts *ProxyHandlerOpts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Host = opts.StorageAddr
		if r.URL.Scheme == "" {
			r.URL.Scheme = "http"
		}
		resp, err := http.DefaultTransport.RoundTrip(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer resp.Body.Close()

		for header, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(header, value)
			}
		}
		w.WriteHeader(resp.StatusCode)

		if _, err := io.Copy(w, resp.Body); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
}

type StaticHandlerOpts struct{}

func StaticHandler(*StaticHandlerOpts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f := r.URL.Path
		if f == "/" {
			f = "/index.html"
		}

		content, err := staticData.ReadFile("static" + f)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=360, immutable")
		w.Header().Set("Content-Type", contentType(f))
		if _, err := w.Write(content); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
}

//go:embed static/*
var staticData embed.FS

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
