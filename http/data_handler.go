package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/s5i/ruuvi2db/data"
)

func dataHandler(db DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		endTime, err := endTime(r, time.Now())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		duration, err := duration(r, 24*time.Hour)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		kind, err := dataKind(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		src := db.Get(endTime.Add(-duration), endTime)
		m := map[time.Time]map[string]any{}
		for _, p := range src {
			if m[p.Timestamp] == nil {
				m[p.Timestamp] = map[string]any{}
			}
			m[p.Timestamp]["ts"] = p.Timestamp
			m[p.Timestamp][data.HumanName(p.Address)] = dataValue(p, kind)
		}
		ret := []map[string]any{}
		for _, v := range m {
			ret = append(ret, v)
		}

		sort.Slice(ret, func(i, j int) bool { return ret[i]["ts"].(time.Time).Before(ret[j]["ts"].(time.Time)) })

		b, err := json.Marshal(ret)
		if err != nil {
			http.Error(w, "Something went wrong", 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if endTime.Before(time.Now()) {
			w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
		}
		if _, err := w.Write(b); err != nil {
			http.Error(w, "Something went wrong", 500)
			return
		}
	}
}

func dataValue(p data.Point, kind string) any {
	switch kind {
	case "temperature":
		return p.Temperature
	case "humidity":
		return p.Humidity
	case "pressure":
		return p.Pressure
	case "battery":
		return p.Battery
	}
	return nil
}

func dataKind(r *http.Request) (string, error) {
	x, ok := r.URL.Query()["kind"]
	if !ok {
		return "", errors.New("kind not specified")
	}

	if len(x) != 1 {
		return "", errors.New("kind specified multiple times")
	}

	return x[0], nil
}
