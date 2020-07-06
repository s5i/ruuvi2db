package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/s5i/ruuvi2db/bluetooth"
	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/db"
	"github.com/s5i/ruuvi2db/db/influx"
	"github.com/s5i/ruuvi2db/db/iowriter"
	"github.com/s5i/ruuvi2db/protocol"
)

var (
	hciID       = flag.Int("hci_device_id", -1, "HCI device to use. -1 probes everything.")
	pointPeriod = flag.Duration("point_period", 5*time.Second, "How often to report a point.")

	humanNames = flag.String("human_names", "", "Comma-separated list of human_name@mac pairs.")

	logToStdout = flag.Bool("log_to_stdout", false, "Whether to log readings to STDOUT.")
	logToInflux = flag.Bool("log_to_influx", false, "Whether to log readings to InfluxDB.")

	debug = flag.Bool("debug", false, "Whether to show debug messages (log package).")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	if !(*debug) {
		log.SetOutput(ioutil.Discard)
	}

	buf := data.NewBuffer(5)

	for _, pair := range strings.Split(*humanNames, ",") {
		if pair == "" {
			continue
		}
		if !strings.Contains(pair, "@") {
			log.Fatalf("human_names contains invalid entry: %q", pair)
		}
		hm := strings.Split(pair, "@")
		data.RegisterHumanName(hm[1], hm[0])
	}

	go func() {
		if err := bluetooth.Run(ctx, *hciID, func(addr string, data []byte) {
			res, err := protocol.ParseDatagram(data, addr)
			if err != nil {
				return
			}
			buf.Push(*res)
		}); err != nil {
			log.Fatalf("bluetooth.Run failed: %v", err)
		}
	}()

	outputs := []db.Interface{}

	if *logToStdout {
		outputs = append(outputs, iowriter.NewStdout())
	}

	if *logToInflux {
		db := influx.NewDB()
		go func() {
			if err := db.RunWithFlagOptions(ctx); err != nil {
				log.Fatalf("influx: db.Run failed: %v", err)
			}
		}()
		outputs = append(outputs, db)
	}

	go func() {
		for now := range time.Tick(*pointPeriod) {
			pts := buf.PullAll(now)
			sort.Slice(pts, func(i, j int) bool {
				return strings.Compare(pts[i].Name(), pts[j].Name()) < 0
			})
			for _, out := range outputs {
				out.Push(pts)
			}
		}
	}()

	select {}
}
