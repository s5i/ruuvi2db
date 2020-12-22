package data

import (
	"sync"
	"time"
)

type Buffer struct {
	mu               sync.RWMutex
	readings         map[string][]Point
	nextIdx          map[string]int
	size             int
	extrapolationGap time.Duration
}

// NewBuffer creates a buffer for data point readings.
// Size specifies how many points should be kept per device address, automatically adjusted to be at least 2.
// ExtrapolationGap specifies the maximum duration between timestamps to be considered for extrapolation. Defaults to 5m.
func NewBuffer(size int, extrapolationGap time.Duration) *Buffer {
	if size < 2 {
		size = 2
	}
	if extrapolationGap == 0 {
		extrapolationGap = 5 * time.Minute
	}
	return &Buffer{
		readings:         map[string][]Point{},
		nextIdx:          map[string]int{},
		size:             size,
		extrapolationGap: extrapolationGap,
	}
}

// Push adds a data point for a given device address.
func (b *Buffer) Push(p Point) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.readings[p.Address] == nil {
		b.readings[p.Address] = make([]Point, b.size, b.size)
		for i := range b.readings[p.Address] {
			b.readings[p.Address][i] = p
		}
	}

	b.readings[p.Address][b.nextIdx[p.Address]] = p
	b.nextIdx[p.Address]++
	b.nextIdx[p.Address] %= b.size
}

// PullAll returns data points for all known devices.
func (b *Buffer) PullAll(timestamp time.Time) []Point {
	b.mu.RLock()
	defer b.mu.RUnlock()

	ret := []Point{}
	for addr := range b.readings {
		if b.readings[addr] == nil {
			continue
		}

		p, err := LinearExtrapolate(b.readings[addr], timestamp, b.extrapolationGap)
		if err != nil {
			continue
		}

		ret = append(ret, p)
	}
	return ret
}
