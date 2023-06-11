package data

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type Buffer struct {
	mu           sync.Mutex
	points       map[string]Point
	maxStaleness time.Duration
}

// NewBuffer creates a buffer for data point readings.
func NewBuffer(maxStaleness time.Duration) *Buffer {
	return &Buffer{
		points:       map[string]Point{},
		maxStaleness: maxStaleness,
	}
}

// Push adds a data point for a given device address.
func (b *Buffer) Push(p Point) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.points[p.Address] = p
}

// PullAll returns data points for all known devices.
func (b *Buffer) PullAll() []Point {
	b.mu.Lock()
	defer b.mu.Unlock()

	ret := []Point{}
	for _, p := range b.points {
		if p.Timestamp.Add(b.maxStaleness).Before(time.Now()) {
			continue
		}

		ret = append(ret, p)
	}

	return ret
}

// Print dumps data points for all known devices to stdout.
func (b *Buffer) Print() {
	b.mu.Lock()
	defer b.mu.Unlock()

	fmt.Fprintln(os.Stdout, "[Buffer dump]:")
	for _, p := range b.points {
		fmt.Fprintf(os.Stdout, "- %s", p)
		if p.Timestamp.Add(b.maxStaleness).Before(time.Now()) {
			fmt.Fprint(os.Stdout, " [stale]")
		}
		fmt.Fprint(os.Stdout, "\n")
	}
}
