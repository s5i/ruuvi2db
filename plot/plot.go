package plot

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/db"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

func NewPlotter(src db.Source) *Plotter {
	return &Plotter{
		format:     "png",
		xSize:      "20cm",
		ySize:      "20cm",
		legendSize: "5cm",
		source:     src,
	}
}

type Plotter struct {
	format     string
	xSize      string
	ySize      string
	legendSize string
	source     db.Source
}

func (z *Plotter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	endTime := time.Now()
	duration := 24 * time.Hour
	var f func(io.Writer, map[string][]data.Point) error

	if x, ok := r.URL.Query()["param"]; ok && len(x) == 1 {
		switch x[0] {
		case "temperature":
			f = z.writeTemperature
		case "humidity":
			f = z.writeHumidity
		case "pressure":
			f = z.writePressure
		case "battery":
			f = z.writeBattery
		default:
			http.Error(w, fmt.Sprintf("bad param %q", x[0]), 400)
			return
		}
	}

	// Unitless end_time is treated as Unix timestamp (seconds).
	// Unitful end_time is treated as offset from time.Now().
	if x, ok := r.URL.Query()["end_time"]; ok && len(x) == 1 {
		if x[0] == "" || strings.ToLower(x[0]) == "now" {
			// Assume default.
		} else if ts, err := strconv.ParseInt(x[0], 10, 64); err == nil {
			endTime = time.Unix(ts, 0)
		} else if dur, err := parseDuration(x[0]); err == nil {
			endTime = time.Now().Add(dur)
		} else {
			http.Error(w, fmt.Sprintf("bad end_time %q", x[0]), 400)
			return
		}
	}

	// Unitless duration is treated as seconds.
	if x, ok := r.URL.Query()["duration"]; ok && len(x) == 1 {
		if x[0] == "" {
			// Assume default.
		} else if durInt, err := strconv.Atoi(x[0]); err == nil {
			duration = time.Duration(durInt) * time.Second
		} else if dur, err := parseDuration(x[0]); err == nil {
			duration = dur
		} else {
			http.Error(w, fmt.Sprintf("bad duration %q", x[0]), 400)
			return
		}
	}

	startTime := endTime.Add(-duration)

	pts := z.source.Get(startTime, endTime)

	w.Header().Set("Content-Type", "image/png")
	if err := f(w, pts); err != nil {
		http.Error(w, "Something went wrong", 500)
		return
	}
}

func (z *Plotter) writeTemperature(w io.Writer, points map[string][]data.Point) error {
	return z.write(w, "Temperature", "°C", points, func(p data.Point) float64 { return p.Temperature })
}

func (z *Plotter) writeHumidity(w io.Writer, points map[string][]data.Point) error {
	return z.write(w, "Humidity", "%", points, func(p data.Point) float64 { return p.Humidity })
}

func (z *Plotter) writePressure(w io.Writer, points map[string][]data.Point) error {
	return z.write(w, "Pressure", "hPa", points, func(p data.Point) float64 { return p.Pressure })
}

func (z *Plotter) writeBattery(w io.Writer, points map[string][]data.Point) error {
	return z.write(w, "Battery", "mV", points, func(p data.Point) float64 { return p.Battery })
}

func (z *Plotter) write(w io.Writer, title, unit string, points map[string][]data.Point, property func(data.Point) float64) error {
	p, err := plot.New()
	if err != nil {
		return err
	}

	p.Title.Text = title
	p.Y.Label.Text = unit
	p.X.Tick.Marker = plot.TimeTicks{Format: time.Stamp, Time: plot.UnixTimeIn(time.Local)}

	lps := []interface{}{}
	for addr, pts := range points {
		lps = append(lps, data.HumanName(addr))
		lps = append(lps, toXYs(pts, property))
	}

	if err := plotutil.AddLines(p, lps...); err != nil {
		return err
	}

	x, err := vg.ParseLength(z.xSize)
	if err != nil {
		return err
	}

	y, err := vg.ParseLength(z.ySize)
	if err != nil {
		return err
	}

	legX, err := vg.ParseLength(z.legendSize)
	if err != nil {
		return err
	}

	wt, err := draw.NewFormattedCanvas(x+legX, y, z.format)
	if err != nil {
		return err
	}

	p.Legend.Top = true
	p.Legend.XOffs += legX
	c := draw.Crop(draw.New(wt), 0, -legX, 0, 0)

	p.Draw(c)

	_, err = wt.WriteTo(w)
	return err
}

func toXYs(points []data.Point, property func(data.Point) float64) plotter.XYs {
	pts := make(plotter.XYs, len(points))
	for i, p := range points {
		pts[i].X = float64(p.Timestamp.Unix())
		pts[i].Y = property(p)
	}
	return pts
}

func parseDuration(x string) (time.Duration, error) {
	if strings.Contains(x, "d") {
		dur, err := time.ParseDuration(strings.ReplaceAll(x, "d", "h"))
		return time.Duration(24) * dur, err
	}
	return time.ParseDuration(x)
}
