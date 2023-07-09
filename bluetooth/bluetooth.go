package bluetooth

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
)

func Run(ctx context.Context, callback func(addr string, mfID uint16, data []byte)) error {
	d, err := linux.NewDevice()
	if err != nil {
		return fmt.Errorf("linux.NewDevice failed: %v", err)
	}

	if err := d.Scan(ctx, false, func(a ble.Advertisement) {
		if len(a.ManufacturerData()) < 2 {
			return
		}
		callback(a.Addr().String(), binary.LittleEndian.Uint16(a.ManufacturerData()[0:2]), a.ManufacturerData()[2:])
	}); err != nil && err != context.Canceled {
		return err
	}
	return nil
}
