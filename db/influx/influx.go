package influx

import (
	"context"
	"flag"
	"fmt"
	"log"

	_ "github.com/influxdata/influxdb1-client"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/db"
)

var (
	influxConnection       = flag.String("influx_connection", "http://localhost:8086", "InfluxDB connection string.")
	influxDatabase         = flag.String("influx_database", "ruuvi", "InfluxDB database.")
	influxTable            = flag.String("influx_table", "ruuvi", "InfluxDB table.")
	influxUsername         = flag.String("influx_username", "", "Username used to connect to InfluxDB.")
	influxPassword         = flag.String("influx_password", "", "Password used to connect to InfluxDB.")
	influxPrecision        = flag.String("influx_precision", "s", "Precision specified when pushing data to InfluxDB.")
	influxRetentionPolicy  = flag.String("influx_retention_policy", "", "Retention policy specified when pushing data to InfluxDB.")
	influxWriteConsistency = flag.String("influx_write_consistency", "", "Write consistency specified when pushing data to InfluxDB.")
)

// NewDB returns an object that can be used to connect and push to Influx DB.
func NewDB() *influxDB {
	return &influxDB{
		dp: make(chan []data.Point),
	}
}

// RunWithFlagOptions starts a connection to DB and handles Push calls.
// The arguments passed onto Run are taken from influx_* flags.
func (idb *influxDB) RunWithFlagOptions(ctx context.Context) error {
	return idb.Run(ctx,
		*influxConnection,
		*influxDatabase,
		*influxTable,
		WithUsername(*influxUsername),
		WithPassword(*influxPassword),
		WithPrecision(*influxPrecision),
		WithRetentionPolicy(*influxRetentionPolicy),
		WithWriteConsistency(*influxWriteConsistency))
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
