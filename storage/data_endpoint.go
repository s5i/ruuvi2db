package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/s5i/ruuvi2db/data"
)

type RunDataEndpointOpts struct {
	Listen       string
	PointsF      func(startTime, endTime time.Time) ([]*data.Point, error)
	AliasF       func(string) (string, error)
	ListAliasesF func() (map[string]string, error)
}

func RunDataEndpoint(ctx context.Context, opts *RunDataEndpointOpts) error {
	srv := http.Server{}
	srv.Addr = opts.Listen

	srv.ReadTimeout = time.Minute
	srv.WriteTimeout = time.Minute
	srv.SetKeepAlivesEnabled(false)

	mux := http.NewServeMux()

	mux.Handle("/data.json", DataHandler(&DataHandlerOpts{
		PointsF: opts.PointsF,
		AliasF:  opts.AliasF,
	}))

	mux.Handle("/aliases.json", AliasesHandler(&AliasesHandlerOpts{
		ListAliasesF: opts.ListAliasesF,
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

type DataHandlerOpts struct {
	PointsF func(startTime, endTime time.Time) ([]*data.Point, error)
	AliasF  func(string) (string, error)
}

func DataHandler(opts *DataHandlerOpts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		endTime, err := dataEndTime(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		duration, err := dataDuration(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		kind, err := dataKind(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		src, err := opts.PointsF(endTime.Add(-duration), endTime)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if endTime.Before(time.Now()) {
			w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
		}

		m := map[time.Time]map[string]any{}
		for _, p := range src {
			if m[p.Timestamp] == nil {
				m[p.Timestamp] = map[string]any{}
			}

			name, err := opts.AliasF(p.Address)
			if err != nil || name == "" {
				name = p.Address
			}

			m[p.Timestamp]["ts"] = p.Timestamp
			m[p.Timestamp][name] = dataValue(p, kind)
		}

		ret := []map[string]any{}
		for _, v := range m {
			ret = append(ret, v)
		}

		sort.Slice(ret, func(i, j int) bool { return ret[i]["ts"].(time.Time).Before(ret[j]["ts"].(time.Time)) })

		e := json.NewEncoder(w)
		e.SetIndent("", "  ")

		if err := e.Encode(ret); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
}

type AliasesHandlerOpts struct {
	ListAliasesF func() (map[string]string, error)
}

func AliasesHandler(opts *AliasesHandlerOpts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		aliases, err := opts.ListAliasesF()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")

		if err := e.Encode(aliases); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
}

func dataValue(p *data.Point, kind string) any {
	switch kind {
	case "temperature":
		return fmt.Sprintf("%.2f", p.Temperature)
	case "humidity":
		return fmt.Sprintf("%.2f", p.Humidity)
	case "pressure":
		return fmt.Sprintf("%.2f", p.Pressure)
	case "battery":
		return fmt.Sprintf("%.2f", p.Battery)
	}
	return nil
}

func dataKind(r *http.Request) (string, error) {
	x, err := singleStringParam(r, "kind")
	if err != nil {
		return "", err
	}

	switch x {
	case "temperature", "humidity", "pressure", "battery":
	default:
		return "", fmt.Errorf("unrecognized kind %q", x)
	}

	return x, nil
}

func dataEndTime(r *http.Request) (time.Time, error) {
	x, err := singleStringParam(r, "end_time")
	if err != nil {
		return time.Time{}, err
	}

	ret, err := strconv.ParseInt(x, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("malformed end_time %q", x)
	}

	return time.Unix(ret, 0), nil
}

func dataDuration(r *http.Request) (time.Duration, error) {
	x, err := singleStringParam(r, "duration")
	if err != nil {
		return 0, err
	}

	ret, err := strconv.ParseInt(x, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("malformed duration %q", x)
	}

	return time.Duration(ret) * time.Second, nil
}
