package reader

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/reader/bluetooth"
	"github.com/s5i/ruuvi2db/reader/protocol"
)

type RunBluetoothOpts struct {
	WatchdogTimeout time.Duration
	CachePointF     func(*data.Point)
}

func RunBluetooth(ctx context.Context, opts *RunBluetoothOpts) error {
	switch err := bluetooth.Run(ctx, func(addr string, mfID uint16, datagram []byte) {
		p, err := protocol.ParseDatagram(mfID, datagram, addr)
		if err != nil {
			return
		}
		opts.CachePointF(p)
	}, opts.WatchdogTimeout); {
	case errors.Is(err, context.Canceled):
		return nil
	case errors.Is(err, bluetooth.ErrInit):
		return fmt.Errorf("%v\nDid you set the capability?\n$ sudo setcap cap_net_raw,cap_net_admin=ep /path/to/reader", err)
	default:
		return err
	}
}
