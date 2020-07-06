package influx

import (
	"context"
	"fmt"
	"log"

	_ "github.com/influxdata/influxdb1-client"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/s5i/ruuvi2db/config"
	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/db"
)

// NewDB returns an object that can be used to connect and push to Influx DB.
func NewDB() *influxDB {
	return &influxDB{
		dp: make(chan []data.Point),
	}
}

// RunWithConfig starts a connection to DB and handles Push calls.
// The arguments passed onto Run are taken from config proto.
func (idb *influxDB) RunWithConfig(ctx context.Context, cfg *config.Config) error {
	opts := []runOption{}

	if u := cfg.GetInfluxDb().Username; u != "" {
		opts = append(opts, WithUsername(u))
	}
	if p := cfg.GetInfluxDb().Password; p != "" {
		opts = append(opts, WithPassword(p))
	}
	if p := cfg.GetInfluxDb().Precision; p != "" {
		opts = append(opts, WithPrecision(p))
	}
	if rp := cfg.GetInfluxDb().RetentionPolicy; rp != "" {
		opts = append(opts, WithRetentionPolicy(rp))
	}
	if wc := cfg.GetInfluxDb().WriteConsistency; wc != "" {
		opts = append(opts, WithWriteConsistency(wc))
	}

	return idb.Run(ctx,
		cfg.GetInfluxDb().Connection,
		cfg.GetInfluxDb().Database,
		cfg.GetInfluxDb().Table,
		opts...)
}

// Run starts a connection to DB and handles Push calls.
func (idb *influxDB) Run(ctx context.Context, connection, database, table string, opts ...runOption) error {
	idb.connection = connection
	idb.database = database
	idb.table = table

	for _, opt := range opts {
		if err := opt(idb); err != nil {
			return fmt.Errorf("run option failed: %v", err)
		}
	}

	c, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:     idb.connection,
		Username: idb.username,
		Password: idb.password,
	})
	if err != nil {
		return fmt.Errorf("influxdb.NewHTTPClient failed: %v", err)
	}
	defer c.Close()

	for {
		select {
		case points := <-idb.dp:
			bp, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
				Database:         idb.database,
				Precision:        idb.precision,
				RetentionPolicy:  idb.retentionPolicy,
				WriteConsistency: idb.writeConsistency,
			})
			if err != nil {
				log.Printf("influxdb.NewBatchPoints failed: %v", err)
				continue
			}
			for _, p := range points {
				pt, err := influxdb.NewPoint(
					idb.table,
					map[string]string{
						"name": p.Name(),
					},
					map[string]interface{}{
						"temperature": p.Temperature,
						"humidity":    p.Humidity,
						"pressure":    p.Pressure,
						"battery":     p.Battery,
					},
					p.Timestamp,
				)
				if err != nil {
					log.Printf("influxdb.NewPoint failed: %v", err)
					continue
				}
				bp.AddPoint(pt)

				if err := c.Write(bp); err != nil {
					log.Printf("c.Write failed: %v", err)
				}
			}

		case <-ctx.Done():
			return nil
		}
	}
}

// Push attempts to send data to DB.
func (ib *influxDB) Push(points []data.Point) {
	ib.dp <- points
}

type influxDB struct {
	dp               chan []data.Point
	connection       string
	database         string
	table            string
	username         string
	password         string
	precision        string
	retentionPolicy  string
	writeConsistency string
}

var _ db.Interface = (*influxDB)(nil)
