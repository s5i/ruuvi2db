package http

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/db"
)

func newCSVHandler(src db.Source) *csvHandler {
	return &csvHandler{
		source: src,
	}
}

type csvHandler struct {
	source db.Source
}

func (z *csvHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	endTime, err := endTime(r, time.Now())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	duration, err := duration(r, time.Hour)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
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

	for addr, pts := range z.source.Get(endTime.Add(-duration), endTime) {
		for _, p := range pts {
			if err := csvW.Write([]string{
				data.HumanName(addr),
				strconv.FormatInt(p.Timestamp.Unix(), 10),
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
