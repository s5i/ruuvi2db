package http

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"github.com/s5i/ruuvi2db/data"
)

func newCSVHandler(src DB) *csvHandler {
	return &csvHandler{
		source: src,
	}
}

type csvHandler struct {
	source DB
}

func (z *csvHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "text/plain")
	if endTime.Before(time.Now()) {
		w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
	}

	csvW := csv.NewWriter(w)
	if err := csvW.Write([]string{
		"name",
		"timestamp",
		"temperature",
		"humidity",
		"pressure",
		"battery",
	}); err != nil {
		http.Error(w, "Something went wrong", 500)
		return
	}

	ptsMap := z.source.Get(endTime.Add(-duration), endTime)

	for addr, pts := range ptsMap {
		for _, p := range pts {
			ts := p.Timestamp.Unix()
			if err := csvW.Write([]string{
				data.HumanName(addr),
				strconv.FormatInt(ts, 10),
				strconv.FormatFloat(p.Temperature, 'f', 2, 64),
				strconv.FormatFloat(p.Humidity, 'f', 2, 64),
				strconv.FormatFloat(p.Pressure, 'f', 2, 64),
				strconv.FormatFloat(p.Battery, 'f', 2, 64),
			}); err != nil {
				http.Error(w, "Something went wrong", 500)
				return
			}
		}
	}
	csvW.Flush()
}
