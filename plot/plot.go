package plot

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/db"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func NewPlotter(src db.Source) *Plotter {
	return &Plotter{
		format: "png",
		xSize:  "20cm",
		ySize:  "20cm",
		source: src,
	}
}

type Plotter struct {
	format string
	xSize  string
	ySize  string
	source db.Source
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

	if x, ok := r.URL.Query()["end_time"]; ok && len(x) == 1 {
		ts, err := strconv.ParseInt(x[0], 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("bad end_time %q", x[0]), 400)
			return
		}
		endTime = time.Unix(ts, 0)

	}
	if x, ok := r.URL.Query()["duration"]; ok && len(x) == 1 {
		dur, err := strconv.Atoi(x[0])
		if err != nil {
			http.Error(w, fmt.Sprintf("bad duration %q", x[0]), 400)
			return
		}
		duration = time.Duration(dur) * time.Second
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

	if err := plotutil.AddLinePoints(p, lps...); err != nil {
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

	wt, err := p.WriterTo(x, y, z.format)
	if err != nil {
		return err
	}

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
