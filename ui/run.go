package ui

import (
	"context"

	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, g *errgroup.Group, cfg *Config) {
	if cfg.ProvidedEndpoints.UI != "" {
		g.Go(func() error {
			return RunUIEndpoint(ctx, &RunUIEndpointOpts{
				Listen:      cfg.ProvidedEndpoints.UI,
				StorageAddr: cfg.ConsumedEndpoints.Storage,
			})
		})
	}
}
