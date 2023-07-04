package bluetooth

import (
	"context"
	"fmt"

	"tinygo.org/x/bluetooth"
)

func Run(ctx context.Context, callback func(addr string, mfID uint16, data []byte)) error {
	adapter := bluetooth.DefaultAdapter
	if err := adapter.Enable(); err != nil {
		return fmt.Errorf("bluetooth.DefaultAdapter.Enable() failed: %v", err)
	}

	go func() {
		<-ctx.Done()
		adapter.StopScan()
	}()

	if err := adapter.Scan(func(a *bluetooth.Adapter, d bluetooth.ScanResult) {
		for k, v := range d.ManufacturerData() {
			callback(d.Address.String(), k, v)
		}
	}); err != nil {
		return err
	}
	return nil
}
