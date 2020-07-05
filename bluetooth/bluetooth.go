package bluetooth

import (
	"context"
	"fmt"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
)

func Run(ctx context.Context, hciID int, callback func(addr string, data []byte)) error {
	opts := []ble.Option{}
	if hciID >= 0 {
		opts = append(opts, ble.OptDeviceID(hciID))
	}
	d, err := linux.NewDevice(opts...)
	if err != nil {
		return fmt.Errorf("linux.NewDevice failed: %v", err)
	}
	return d.Scan(ctx, false,
		func(a ble.Advertisement) {
			callback(a.Addr().String(), a.ManufacturerData())
		})
}
