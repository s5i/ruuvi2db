package bluetooth

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
)

func Run(ctx context.Context, callback func(addr string, mfID uint16, data []byte), watchdogTimeout time.Duration) error {
	d, err := linux.NewDevice()
	if err != nil {
		return fmt.Errorf("linux.NewDevice failed: %v", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	defer wg.Wait()

	watchdogCh := make(chan bool)
	watchdogErr := false
	if watchdogTimeout != 0 {
		wg.Add(1)

		t := time.NewTimer(watchdogTimeout)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-watchdogCh:
					if !t.Stop() {
						<-t.C
					}
					t.Reset(watchdogTimeout)
				case <-t.C:
					watchdogErr = true
					cancel()
					return
				}
			}
		}()
	}

	if err := d.Scan(ctx, false, func(a ble.Advertisement) {
		if watchdogTimeout != 0 {
			watchdogCh <- true
		}

		if len(a.ManufacturerData()) < 2 {
			return
		}
		callback(a.Addr().String(), binary.LittleEndian.Uint16(a.ManufacturerData()[0:2]), a.ManufacturerData()[2:])
	}); err != nil && err != context.Canceled {
		return err
	}

	if watchdogErr {
		return fmt.Errorf("watchdog failure: no advertisement processed for more than %v", watchdogTimeout)
	}
	return nil
}
