package bolt

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/s5i/ruuvi2db/config"
	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/db"
)

// NewDB returns an object that can be used to connect and push to SQLite DB.
func NewDB() *boltDB {
	return &boltDB{
		dp:         make(chan []data.Point, 1),
		bucketSize: 24 * time.Hour,
		retention:  7 * 24 * time.Hour,
	}
}

// RunWithConfig starts a connection to DB and handles Push calls.
// The arguments passed onto Run are taken from config proto.
func (bdb *boltDB) RunWithConfig(ctx context.Context, cfg *config.Config) error {
	opts := []runOption{}

	if r := cfg.GetBoltDb().RetentionSec; r != 0 {
		opts = append(opts, WithRetention(time.Duration(r)*time.Second))
	}
	if bs := cfg.GetBoltDb().BucketSizeSec; bs != 0 {
		opts = append(opts, WithBucketSize(time.Duration(bs)*time.Second))
	}

	return bdb.Run(ctx, cfg.GetBoltDb().Path, opts...)
}

// Run starts a connection to DB and handles Push calls.
func (bdb *boltDB) Run(ctx context.Context, path string, opts ...runOption) error {
	bdb.dbPath = path

	for _, opt := range opts {
		if err := opt(bdb); err != nil {
			return fmt.Errorf("run option failed: %v", err)
		}
	}

	db, err := bolt.Open(path, 0644, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return err
	}
	defer db.Close()

	for {
		select {
		case points := <-bdb.dp:
			for _, p := range points {
				db.Update(func(tx *bolt.Tx) error {
					b, err := tx.CreateBucketIfNotExists(bucketFromTimestamp(p.Timestamp, bdb.bucketSize))
					if err != nil {
						return err
					}

					seq, err := b.NextSequence()
					if err != nil {
						return err
					}
					k := keyFromSeq(seq)

					v, err := p.Encode()
					if err != nil {
						return err
					}

					return b.Put(k, v)
				})
			}

		case <-ctx.Done():
			return nil
		}
	}

	return nil
}

// Push attempts to send data to DB.
func (bdb *boltDB) Push(points []data.Point) {
	bdb.dp <- points
}

type boltDB struct {
	dp         chan []data.Point
	dbPath     string
	bucketSize time.Duration
	retention  time.Duration
}

var _ db.Interface = (*boltDB)(nil)

func bucketFromTimestamp(timestamp time.Time, truncate time.Duration) []byte {
	ts := uint64(timestamp.Truncate(truncate).UnixNano())
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, ts)
	return b
}

func keyFromSeq(seq uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, seq)
	return b
}
