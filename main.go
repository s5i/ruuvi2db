package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/paypal/gatt"
	"github.com/s5i/ruuvi2db/bluetooth"
	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/protocol"
)

var (
	hciID       = flag.Int("hci_device_id", -1, "HCI device to use. -1 probes everything.")
	pointPeriod = flag.Duration("point_period", 5*time.Second, "How often to report a point.")
	logToStdout = flag.Bool("log_to_stdout", false, "Whether to log readings to STDOUT.")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	buf := data.NewBuffer(5)

	go func() {
		if err := bluetooth.Run(ctx, *hciID, func(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
			res, err := protocol.ParseDatagram(a.ManufacturerData, p.ID())
			if err != nil {
				return
			}
			buf.Push(*res)
		}); err != nil {
			log.Fatalf("bluetooth.Run failed: %v", err)
		}
	}()

	go func() {
		for now := range time.Tick(*pointPeriod) {
			pts := buf.PullAll(now)
			sort.Slice(pts, func(i, j int) bool {
				return strings.Compare(pts[i].Address, pts[j].Address) < 0
			})
			if *logToStdout {
				fmt.Println("---")
				for _, p := range pts {
					fmt.Println(p)
				}
			}
		}
	}()

	select {}
}
