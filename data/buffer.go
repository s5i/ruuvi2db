package data

import (
	"fmt"
	"os"
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
func NewBuffer() *Buffer {

	return &Buffer{
		readings:         map[string][]Point{},
		nextIdx:          map[string]int{},
		size:             2,
		extrapolationGap: 5 * time.Minute,
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

// Print dumps data points for all known devices to stdout.
func (b *Buffer) Print() {
	b.mu.RLock()
	defer b.mu.RUnlock()

	fmt.Fprintln(os.Stdout, "[Buffer dump]:")
	for addr := range b.readings {
		if b.readings[addr] == nil {
			continue
		}

		p, err := LinearExtrapolate(b.readings[addr], time.Now(), b.extrapolationGap)
		if err != nil {
			continue
		}

		fmt.Fprintf(os.Stdout, "- %s\n", p)
	}
}
