package http

import (
	"embed"
	"fmt"
	"net/http"
)

//go:embed static/*
var StaticData embed.FS

func staticHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f := r.URL.Path
		if f == "/" {
			f = "/index.html"
		}

		content, err := StaticData.ReadFile("static" + f)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("%q, %+v", "static"+f, StaticData)))
			http.Error(w, "Not found", 404)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=360, immutable")
		w.Header().Set("Content-Type", contentType(f))
		if _, err := w.Write(content); err != nil {
			http.Error(w, "Something went wrong", 500)
			return
		}
	}
}
