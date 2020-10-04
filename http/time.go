package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/s5i/ruuvi2db/util"
)

func endTime(r *http.Request, def time.Time) (time.Time, error) {
	x, ok := r.URL.Query()["end_time"]
	if !ok {
		return def, nil
	}

	if len(x) != 1 {
		return time.Time{}, errors.New("end_time specified multiple times")
	}

	t := x[0]

	if strings.ToLower(t) == "now" {
		return time.Now(), nil
	}

	if t == "" {
		return def, nil
	}

	if ts, err := strconv.ParseInt(t, 10, 64); err == nil {
		return time.Unix(ts, 0), nil
	}

	if dur, err := util.ParseDuration(t); err == nil {
		return time.Now().Add(dur), nil
	}

	return time.Time{}, fmt.Errorf("malformed end_time: %v", t)
}

func duration(r *http.Request, def time.Duration) (time.Duration, error) {
	x, ok := r.URL.Query()["duration"]
	if !ok {
		return def, nil
	}

	if len(x) != 1 {
		return 0, errors.New("duration specified multiple times")
	}

	d := x[0]

	if d == "" {
		return def, nil
	}

	if ts, err := strconv.ParseInt(d, 10, 64); err == nil {
		return time.Duration(ts) * time.Second, nil
	}

	if dur, err := util.ParseDuration(d); err == nil {
		return dur, nil
	}

	return 0, fmt.Errorf("malformed duration: %v", d)
}
