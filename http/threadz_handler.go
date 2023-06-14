package http

import (
	"net/http"
	"runtime"
)

func threadzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for len := 8192; true; len *= 2 {
			buf := make([]byte, len)
			written := runtime.Stack(buf, true)
			if written == len {
				continue
			}
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Content-Type", "text/plain")
			if _, err := w.Write(buf[:written]); err != nil {
				http.Error(w, "Something went wrong", 500)
			}
			return
		}
	}
}
