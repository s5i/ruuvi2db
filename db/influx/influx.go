package influx

import (
	"context"
	"fmt"
	"log"

	_ "github.com/influxdata/influxdb1-client"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/s5i/ruuvi2db/data"
)

// NewDB returns an object that can be used to connect and push to Influx DB.
func NewDB() *influxDB {
	return &influxDB{
		dp: make(chan []data.Point),
	}
}

// WithUsername causes Run to use a given username to connect to DB.
func WithUsername(u string) runOption {
	return func(idb *influxDB) error {
		idb.username = u
		return nil
	}
}

// WithPassword causes Run to use a given password to connect to DB.
func WithPassword(p string) runOption {
	return func(idb *influxDB) error {
		idb.password = p
		return nil
	}
}

// WithPrecision causes Run to use a given precision when pushing to DB (default is seconds).
func WithPrecision(p string) runOption {
	return func(idb *influxDB) error {
		idb.precision = p
		return nil
	}
}

// WithRetentionPolicy causes Run to use a given retention policy when pushing to DB.
func WithRetentionPolicy(rp string) runOption {
	return func(idb *influxDB) error {
		idb.retentionPolicy = rp
		return nil
	}
}

// WithWriteConsistency causes Run to use a given write consistency when pushing to DB.
func WithWriteConsistency(wc string) runOption {
	return func(idb *influxDB) error {
		idb.writeConsistency = wc
		return nil
	}
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

type runOption func(*influxDB) error
