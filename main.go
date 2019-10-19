package main

import (
	"context"
	"flag"
	"log"

	"github.com/paypal/gatt"
	"github.com/s5i/ruuvi2db/bluetooth"
)

var (
	hciID = flag.Int("hci_device_id", -1, "HCI device to use. -1 probes everything.")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	go func() {
		if err := bluetooth.Run(ctx, *hciID, func(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
			// TODO(s5i): implement handler; the data lives in a.ManufacturerData and p.ID()
		}); err != nil {
			log.Fatalf("bluetooth.Run failed: %v", err)
		}
	}()

	select {}
}
