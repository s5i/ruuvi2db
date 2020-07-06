package db

import "github.com/s5i/ruuvi2db/data"

type Interface interface {
	Push(points []data.Point)
}
