package iowriter

import "os"

// NewStdout returns an object that outputs Ruuvi data to stdout.
func NewStdout() *iowriter {
	return New(os.Stdout)
}
