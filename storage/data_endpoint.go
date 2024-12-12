package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
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

		resolution, err := dataResolution(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		src, err := opts.PointsF(endTime.Add(-duration), endTime)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		src = align(src, resolution, 2*resolution)

		w.Header().Set("Content-Type", "application/json")

		if endTime.Before(time.Now()) {
			w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
		}

		m := map[time.Time]map[string]any{}
		for _, p := range src {
			if m[p.Timestamp] == nil {
				m[p.Timestamp] = map[string]any{}
			}

			m[p.Timestamp]["ts"] = p.Timestamp
			m[p.Timestamp][p.Address] = dataValue(p, kind)
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

var kinds = []string{"temperature", "humidity", "pressure", "battery"}

func dataKind(r *http.Request) (string, error) {
	x, ok, err := singleStringParam(r, "kind")
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("kind not specified")
	}

	if !slices.Contains(kinds, x) {
		return "", fmt.Errorf("unrecognized kind %q; valid: %q", x, kinds)
	}

	return x, nil
}

func dataEndTime(r *http.Request) (time.Time, error) {
	x, ok, err := singleStringParam(r, "end_time")
	if err != nil {
		return time.Time{}, err
	}
	if !ok {
		return time.Now(), nil
	}

	ret, err := strconv.ParseInt(x, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("malformed end_time %q", x)
	}

	return time.Unix(ret, 0), nil
}

func dataDuration(r *http.Request) (time.Duration, error) {
	x, ok, err := singleStringParam(r, "duration")
	if err != nil {
		return 0, err
	}
	if !ok {
		return time.Hour, nil
	}

	ret, err := strconv.ParseInt(x, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("malformed duration %q", x)
	}

	return time.Duration(ret) * time.Second, nil
}

func dataResolution(r *http.Request) (time.Duration, error) {
	x, ok, err := singleStringParam(r, "resolution")
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("empty resolution")
	}

	ret, err := strconv.ParseInt(x, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("malformed resolution %q", x)
	}

	return time.Duration(ret) * time.Second, nil
}

func align(src []*data.Point, resolution time.Duration, maxGap time.Duration) []*data.Point {
	if resolution == 0 {
		return src
	}
	raw := map[string][]*data.Point{}
	for _, p := range src {
		raw[p.Address] = append(raw[p.Address], p)
	}
	for addr := range raw {
		slices.SortFunc(raw[addr], func(a, b *data.Point) int {
			return int(a.Timestamp.Unix() - b.Timestamp.Unix())
		})
	}
	sorted := raw
	aligned := map[string][]*data.Point{}

	for addr, pts := range sorted {
		if len(pts) == 0 {
			continue
		}

		get := func(i int) (*data.Point, bool) {
			if i < 0 || i >= len(pts) {
				return nil, false
			}
			return pts[i], true
		}

		for i, outTS, maxTS := 0, pts[0].Timestamp.Truncate(resolution), pts[len(pts)-1].Timestamp; ; {
			left, leftOk := get(i)
			right, rightOk := get(i + 1)

			// left and right are two neighboring points.
			// outTS is the timestamp at which we'd like to output a point.
			// We can output a point if either:
			// - we have an exact timestamp match
			// - outTS lies between left and right and they're not spread apart more than maxGap

			// Basic checks: no more points / output TS out of range.
			if outTS.After(maxTS) {
				break
			}

			// Skip output TS if it'd fall to the left of interpolation window.
			if leftOk && outTS.Before(left.Timestamp) {
				outTS = outTS.Add(resolution)
				continue
			}

			// Skip input point if the desired timestamp would fall to the right of interpolation window.
			if rightOk && right.Timestamp.Before(outTS) {
				i++
				continue
			}

			// Exact match; output point, move to next output TS.
			if left.Timestamp.Equal(outTS) {
				aligned[addr] = append(aligned[addr], left)
				outTS = outTS.Add(resolution)
				continue
			}

			// No more "right" points, can't perform any more interpolations (and we already checked exact match).
			if !rightOk {
				break
			}

			// Interpolation window too wide, skip to the next one.
			if right.Timestamp.Sub(left.Timestamp) > maxGap {
				i++
				continue
			}

			// outTS lies in a narrow enough interpolation window.
			// Calculate linear extrapolation coefficient.
			// 0 if outTS == left.Timestamp, 1 if outTS == right.Timestamp
			// out.Field = left.Field + coeff * (right.Field - left.Field)
			coeff := float64(outTS.Sub(left.Timestamp)) / float64(right.Timestamp.Sub(left.Timestamp))

			aligned[addr] = append(aligned[addr], &data.Point{
				Address:     left.Address,
				Timestamp:   outTS,
				Temperature: left.Temperature + coeff*(right.Temperature-left.Temperature),
				Humidity:    left.Humidity + coeff*(right.Humidity-left.Humidity),
				Pressure:    left.Pressure + coeff*(right.Pressure-left.Pressure),
				Battery:     left.Battery + coeff*(right.Battery-left.Battery),
			})
			outTS = outTS.Add(resolution)
		}
	}

	var ret []*data.Point
	for _, pts := range aligned {
		ret = append(ret, pts...)
	}
	return ret
}
