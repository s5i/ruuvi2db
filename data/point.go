package data

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"strings"
	"time"
)

// Point represents a single data point from a RuuviTag.
type Point struct {
	Address   string    `json:",omitempty"`
	Timestamp time.Time `json:",omitempty"`

	Temperature float64 `json:",omitempty"`
	Humidity    float64 `json:",omitempty"`
	Pressure    float64 `json:",omitempty"`
	Battery     float64 `json:",omitempty"`
}

func (d Point) String() string {
	return fmt.Sprintf("%s @ %v: %.2f °C, %.2f%% humid, %.2f hPa, %.2f mV", d.Address, d.Timestamp.Format(time.DateTime), d.Temperature, d.Humidity, d.Pressure, d.Battery)
}

func (d Point) Encode() ([]byte, error) {
	mac, err := net.ParseMAC(d.Address)
	if err != nil {
		return nil, err
	}

	b := make([]byte, 46)
	binary.BigEndian.PutUint64(b[0:8], uint64(d.Timestamp.UnixNano()))
	binary.BigEndian.PutUint64(b[8:16], math.Float64bits(d.Temperature))
	binary.BigEndian.PutUint64(b[16:24], math.Float64bits(d.Humidity))
	binary.BigEndian.PutUint64(b[24:32], math.Float64bits(d.Pressure))
	binary.BigEndian.PutUint64(b[32:40], math.Float64bits(d.Battery))
	copy(b[40:46], mac)

	return b, nil
}

func DecodePoint(b []byte) (*Point, error) {
	if got, want := len(b), 46; got != want {
		return nil, fmt.Errorf("got %d bytes, want %d", got, want)
	}
	return &Point{
		Timestamp:   time.Unix(0, int64(binary.BigEndian.Uint64(b[0:8]))),
		Temperature: math.Float64frombits(binary.BigEndian.Uint64(b[8:16])),
		Humidity:    math.Float64frombits(binary.BigEndian.Uint64(b[16:24])),
		Pressure:    math.Float64frombits(binary.BigEndian.Uint64(b[24:32])),
		Battery:     math.Float64frombits(binary.BigEndian.Uint64(b[32:40])),
		Address:     strings.ToUpper(net.HardwareAddr(b[40:46]).String()),
	}, nil
}
