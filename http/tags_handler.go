package http

import (
	"encoding/json"
	"net/http"

	"github.com/s5i/ruuvi2db/data"
	"golang.org/x/exp/slices"
)

func tagsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tags := data.ListHumanNames()
		slices.Sort(tags)
		b, err := json.Marshal(tags)
		if err != nil {
			http.Error(w, "Something went wrong", 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(b); err != nil {
			http.Error(w, "Something went wrong", 500)
			return
		}
	}
}
