package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/s5i/ruuvi2db/bluetooth"
	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/db/influx"
	"github.com/s5i/ruuvi2db/protocol"
)

var (
	hciID       = flag.Int("hci_device_id", -1, "HCI device to use. -1 probes everything.")
	pointPeriod = flag.Duration("point_period", 5*time.Second, "How often to report a point.")

	humanNames = flag.String("human_names", "", "Comma-separated list of human_name@mac pairs.")

	logToStdout = flag.Bool("log_to_stdout", false, "Whether to log readings to STDOUT.")

	logToInflux            = flag.Bool("log_to_influx", false, "Whether to log readings to InfluxDB.")
	influxConnection       = flag.String("influx_connection", "http://localhost:8086", "InfluxDB connection string.")
	influxDatabase         = flag.String("influx_database", "ruuvi", "InfluxDB database.")
	influxTable            = flag.String("influx_table", "ruuvi", "InfluxDB table.")
	influxUsername         = flag.String("influx_username", "", "Username used to connect to InfluxDB.")
	influxPassword         = flag.String("influx_password", "", "Password used to connect to InfluxDB.")
	influxPrecision        = flag.String("influx_precision", "s", "Precision specified when pushing data to InfluxDB.")
	influxRetentionPolicy  = flag.String("influx_retention_policy", "", "Retention policy specified when pushing data to InfluxDB.")
	influxWriteConsistency = flag.String("influx_write_consistency", "", "Write consistency specified when pushing data to InfluxDB.")

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

	type output interface {
		Push(points []data.Point)
	}
	outputs := []output{}

	if *logToStdout {
		outputs = append(outputs, &stdout{})
	}

	if *logToInflux {
		db := influx.NewDB()
		go func() {
			if err := db.Run(ctx, *influxConnection, *influxDatabase, *influxTable,
				influx.WithUsername(*influxUsername),
				influx.WithPassword(*influxPassword),
				influx.WithPrecision(*influxPrecision),
				influx.WithRetentionPolicy(*influxRetentionPolicy),
				influx.WithWriteConsistency(*influxWriteConsistency)); err != nil {
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

type stdout struct{}

func (s *stdout) Push(points []data.Point) {
	fmt.Println("---")
	for _, p := range points {
		fmt.Println(p)
	}
}
