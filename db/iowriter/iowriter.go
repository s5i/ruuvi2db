package iowriter

import (
	"fmt"
	"io"

	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/db"
)

// New returns an object that outputs Ruuvi data to a given io.Writer.
func New(writer io.Writer) *iowriter {
	return &iowriter{
		writer: writer,
	}
}

func (i *iowriter) Push(points []data.Point) {
	fmt.Fprintln(i.writer, "---")
	for _, p := range points {
		fmt.Fprintln(i.writer, p)
	}
	if len(points) == 0 {
		fmt.Fprintln(i.writer, "<no data>")
	}
}

type iowriter struct {
	writer io.Writer
}

var _ db.Interface = (*iowriter)(nil)
