package storage

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/s5i/ruuvi2db/data"
)

func TestAlign(t *testing.T) {
	p := func(t, v int) *data.Point {
		return &data.Point{
			Timestamp:   time.Unix(int64(t), 0),
			Temperature: float64(v),
		}
	}
	mkPoints := func(ts ...int) []*data.Point {
		var ret []*data.Point
		for _, t := range ts {
			ret = append(ret, p(t, t))
		}
		return ret
	}

	for _, tc := range []struct {
		name       string
		in         []*data.Point
		resolution time.Duration
		maxGap     time.Duration
		want       []*data.Point
	}{
		{
			name:       "basic",
			in:         mkPoints(101, 201),
			resolution: 100 * time.Second,
			maxGap:     100 * time.Second,
			want:       mkPoints(200),
		},
		{
			name:       "edge left",
			in:         mkPoints(-100, 101, 201),
			resolution: 100 * time.Second,
			maxGap:     100 * time.Second,
			want:       mkPoints(-100, 200),
		},
		{
			name:       "edge right",
			in:         mkPoints(101, 201, 500),
			resolution: 100 * time.Second,
			maxGap:     100 * time.Second,
			want:       mkPoints(200, 500),
		},
		{
			name:       "center spots",
			in:         mkPoints(101, 201, 500, 700, 900, 1101, 1301),
			resolution: 100 * time.Second,
			maxGap:     100 * time.Second,
			want:       mkPoints(200, 500, 700, 900),
		},
		{
			name:       "gaps_wide",
			in:         mkPoints(101, 201, 501, 701, 901, 1101, 1301),
			resolution: 100 * time.Second,
			maxGap:     100 * time.Second,
			want:       mkPoints(200),
		},
		{
			name:       "gaps_ok",
			in:         mkPoints(101, 201, 301, 501, 701),
			resolution: 100 * time.Second,
			maxGap:     200 * time.Second,
			want:       mkPoints(200, 300, 400, 500, 600, 700),
		},
		{
			name:       "highres",
			in:         mkPoints(101, 201, 301, 501, 701),
			resolution: 50 * time.Second,
			maxGap:     100 * time.Second,
			want:       mkPoints(150, 200, 250, 300),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if diff := cmp.Diff(tc.want, align(tc.in, tc.resolution, tc.maxGap), cmpopts.EquateApprox(0, 0.5), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("align diff -want +got\n%v", diff)
			}
		})
	}
}
