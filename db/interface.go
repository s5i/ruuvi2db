package db

import (
	"time"

	"github.com/s5i/ruuvi2db/data"
)

type Interface interface {
	Push(points []data.Point)
}

type Source interface {
	Get(startTime, endTime time.Time) map[string][]data.Point
}
