package reader

import (
	"context"

	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, g *errgroup.Group, cfg *Config) {
	get, put := MakeCache(&MakeCacheOpts{
		MaxStaleness: cfg.Data.MaxStaleness,
		MACFilter:    cfg.Data.MACFilter,
	})

	g.Go(func() error {
		return RunBluetooth(ctx, &RunBluetoothOpts{
			WatchdogTimeout: cfg.Bluetooth.WatchdogTimeout,
			CachePointF:     put,
		})
	})

	g.Go(func() error {
		return RunDataEndpoint(ctx, &RunDataEndpointOpts{
			Listen:  cfg.ProvidedEndpoints.Data,
			PointsF: get,
		})
	})
}
