package data

import (
	"fmt"
	"sync"
	"time"
)

type Buffer struct {
	mu       sync.RWMutex
	readings map[string][]Point
	nextIdx  map[string]int
	size     int
}

// NewBuffer creates a buffer for data point readings.
// Size specifies how many points should be kept per device address, automatically adjusted to be at least 2.
func NewBuffer(size int) *Buffer {
	if size < 2 {
		size = 2
	}
	return &Buffer{
		readings: map[string][]Point{},
		nextIdx:  map[string]int{},
		size:     size,
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

// Pull returns a data point for a given device address.
func (b *Buffer) Pull(addr string, timestamp time.Time) (Point, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.readings[addr] == nil {
		return Point{}, fmt.Errorf("no readings for device %s", addr)
	}

	p, err := LinearExtrapolate(b.readings[addr], timestamp)
	if err != nil {
		return Point{}, fmt.Errorf("failed to extrapolate readings for device %s: %v", addr, err)
	}

	return p, nil
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

		p, err := LinearExtrapolate(b.readings[addr], timestamp)
		if err != nil {
			continue
		}

		ret = append(ret, p)
	}
	return ret
}
