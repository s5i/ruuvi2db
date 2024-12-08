package storage

import (
	"context"

	"github.com/s5i/ruuvi2db/storage/database/bolt"
	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, g *errgroup.Group, cfg *Config) {
	db := bolt.New()

	g.Go(func() error {
		return db.Run(ctx, &bolt.Config{
			Path:              cfg.Database.Bolt.Path,
			RetentionWindow:   cfg.Database.Bolt.RetentionWindow,
			AllowSchemaUpdate: cfg.Database.Bolt.AllowSchemaUpdate,
		})
	})

	if cfg.ConsumedEndpoints.Reader != "" {
		g.Go(func() error {
			return RunReaderConsumer(ctx, &RunReaderConsumerOpts{
				ReaderAddr:   cfg.ConsumedEndpoints.Reader,
				QueryPeriod:  cfg.ReaderConsumer.QueryPeriod,
				MaxStaleness: cfg.ReaderConsumer.MaxStaleness,
				MACFilter:    cfg.ReaderConsumer.MACFilter,
				PushPointsF:  db.PushPoints,
			})
		})
	}

	if cfg.ProvidedEndpoints.Data != "" {
		g.Go(func() error {
			return RunDataEndpoint(ctx, &RunDataEndpointOpts{
				Listen:       cfg.ProvidedEndpoints.Data,
				PointsF:      db.Points,
				AliasF:       db.Alias,
				ListAliasesF: db.ListAliases,
			})
		})
	}

	if cfg.ProvidedEndpoints.Admin != "" {
		g.Go(func() error {
			return RunAdminEndpoint(ctx, &RunAdminEndpointOpts{
				Listen:    cfg.ProvidedEndpoints.Admin,
				SetAliasF: db.SetAlias,
			})
		})
	}
}
