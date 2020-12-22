package data

import (
	"errors"
	"sort"
	"time"
)

// LinearExtrapolate returns a result of linear extrapolation (or interpolation) of points
// using timestamps as the function domain.
// Warning: no guarantees on returned Address if input Addresses aren't consistent.
func LinearExtrapolate(p []Point, wantTimestamp time.Time, maxExtrapolationGap time.Duration) (Point, error) {
	pts := []Point{}
	minTs := wantTimestamp.Add(-maxExtrapolationGap)
	maxTs := wantTimestamp.Add(maxExtrapolationGap)

	// Get rid of entries with duplicate timestamps, we can't use them.
	seenTs := map[time.Time]bool{}
	for _, p := range p {
		if seenTs[p.Timestamp] {
			continue
		}
		if p.Timestamp.Before(minTs) || p.Timestamp.After(maxTs) {
			continue
		}
		pts = append(pts, p)
		seenTs[p.Timestamp] = true
	}

	if len(pts) < 2 {
		return Point{}, errors.New("need at least 2 points with unique timestamps within the extrapolation gap")
	}

	sort.Slice(pts, func(i, j int) bool {
		a, b, w := pts[i].Timestamp.Unix(), pts[j].Timestamp.Unix(), wantTimestamp.Unix()
		abs := func(x int64) int64 {
			if x > 0 {
				return x
			}
			return -x
		}
		return abs(w-a) < abs(w-b)
	})

	a, b := pts[0], pts[1]
	if b.Timestamp.Before(a.Timestamp) {
		a, b = b, a
	}

	// Linear extrapolation coefficient.
	// 0 if wantTimestamp == a, 1 if wantTimestamp == b
	coeff := (float64(wantTimestamp.Unix()) - float64(a.Timestamp.Unix())) / (float64(b.Timestamp.Unix()) - float64(a.Timestamp.Unix()))

	// res.Field = a.Field + coeff * (b.Field - a.Field)
	return Point{
		Address:     a.Address,
		Timestamp:   wantTimestamp,
		Temperature: a.Temperature + coeff*(b.Temperature-a.Temperature),
		Humidity:    a.Humidity + coeff*(b.Humidity-a.Humidity),
		Pressure:    a.Pressure + coeff*(b.Pressure-a.Pressure),
		Battery:     a.Battery + coeff*(b.Battery-a.Battery),
	}, nil
}
