package data

import (
	"fmt"
	"time"
)

// Point represents a single data point from a RuuviTag.
type Point struct {
	Address   string
	Timestamp time.Time

	Temperature float64
	Humidity    float64
	Pressure    float64
	Battery     float64
}

func (d Point) String() string {
	return fmt.Sprintf("%s: (%.2f °C, %.2f%% humid, %.2f hPa)", d.Address, d.Temperature, d.Humidity, d.Pressure)
}
