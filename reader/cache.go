package reader

import (
	"strings"
	"sync"
	"time"

	"github.com/s5i/ruuvi2db/data"
)

type MakeCacheOpts struct {
	MaxStaleness time.Duration
	MACFilter    []string
}

func MakeCache(opts *MakeCacheOpts) (get func() []*data.Point, put func(*data.Point)) {
	filter := map[string]bool{}
	for _, m := range opts.MACFilter {
		filter[strings.ToUpper(m)] = true
	}

	var mu sync.Mutex
	points := map[string]*data.Point{}

	getF := func() []*data.Point {
		mu.Lock()
		defer mu.Unlock()

		var ret []*data.Point
		for k, v := range points {
			if v.Timestamp.Add(opts.MaxStaleness).Before(time.Now()) {
				delete(points, k)
				continue
			}
			ret = append(ret, v)
		}
		return ret
	}

	putF := func(p *data.Point) {
		mu.Lock()
		defer mu.Unlock()

		if len(filter) == 0 || filter[p.Address] {
			points[p.Address] = p
		}
	}
	return getF, putF
}
