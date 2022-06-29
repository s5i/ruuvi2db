package http

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/s5i/ruuvi2db/config"
	"github.com/s5i/ruuvi2db/data"
)

func newCSVHandler(src DB, cfg *config.Config) *csvHandler {
	return &csvHandler{
		source:                src,
		defaultTimestampLimit: int(cfg.HTTP.DefaultTimestampLimit),
	}
}

type csvHandler struct {
	source                DB
	defaultTimestampLimit int
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
	limit, err := csvLimit(r, z.defaultTimestampLimit)
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

	ptsMap := z.source.Get(endTime.Add(-duration), endTime)
	ltd := csvLimitedTimestamps(ptsMap, limit)

	for addr, pts := range ptsMap {
		for _, p := range pts {
			ts := p.Timestamp.Unix()
			if ltd != nil && !ltd[ts] {
				continue
			}
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

func csvLimitedTimestamps(ptsMap map[string][]data.Point, limit int) map[int64]bool {
	if limit == 0 {
		return nil
	}

	all := map[int64]bool{}
	for _, pts := range ptsMap {
		for _, p := range pts {
			all[p.Timestamp.Unix()] = true
		}
	}

	sorted := make([]int64, 0, len(all))
	for ts := range all {
		sorted = append(sorted, ts)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] > sorted[j]
	})

	out := map[int64]bool{}
	i := 0.0
	skip := float64(len(sorted)) / float64(limit)
	for {
		idx := int(i)
		if idx >= len(sorted) {
			break
		}
		out[sorted[idx]] = true
		i += skip
	}
	return out
}

func csvLimit(r *http.Request, def int) (int, error) {
	x, ok := r.URL.Query()["limit"]
	if !ok {
		return def, nil
	}

	if len(x) != 1 {
		return 0, errors.New("limit specified multiple times")
	}

	if limit, err := strconv.ParseInt(x[0], 10, 64); err == nil {
		return int(limit), nil
	}

	return 0, fmt.Errorf("malformed limit: %v", x)
}
