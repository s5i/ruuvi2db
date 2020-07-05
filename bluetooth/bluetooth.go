package bluetooth

import (
	"context"
	"fmt"

	"github.com/paypal/gatt"
)

func Run(ctx context.Context, hciID int, callback func(addr string, data []byte)) error {
	d, err := gatt.NewDevice(gatt.LnxMaxConnections(1), gatt.LnxDeviceID(hciID, true))
	if err != nil {
		return fmt.Errorf("gatt.NewDevice failed: %v", err)
	}

	d.Handle(gatt.PeripheralDiscovered(func(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
		callback(p.ID(), a.ManufacturerData)
	}))

	powerOn := make(chan gatt.Device)
	powerOff := make(chan gatt.Device)

	d.Init(func(d gatt.Device, s gatt.State) {
		if s == gatt.StatePoweredOn {
			powerOn <- d
		}
		if s == gatt.StatePoweredOff {
			powerOff <- d
		}
	})

	for {
		select {
		case d := <-powerOn:
			d.Scan(nil, true)
		case d := <-powerOff:
			d.StopScanning()
		case <-ctx.Done():
			return nil
		}
	}
}
