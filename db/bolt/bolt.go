package bolt

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/boltdb/bolt"
	"github.com/s5i/ruuvi2db/config"
	"github.com/s5i/ruuvi2db/data"
	"github.com/s5i/ruuvi2db/db"
)

// NewDB returns an object that can be used to connect and push to Bolt DB.
func NewDB() *boltDB {
	return &boltDB{
		pushCh:     make(chan []data.Point, 1),
		getCh:      make(chan getReq),
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

	tick := time.NewTicker(time.Minute)
	defer tick.Stop()

	for {
		select {
		case points := <-bdb.pushCh:
			for _, p := range points {
				if err := db.Update(func(tx *bolt.Tx) error {
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
				}); err != nil {
					log.Printf("db.Update failed: %v", err)
				}
			}

		case getReq := <-bdb.getCh:

			points := []data.Point{}
			db.View(func(tx *bolt.Tx) error {
				return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
					bStart := timestampFromBucket(name)
					bEnd := bStart.Add(bdb.bucketSize)
					if bStart.After(getReq.endTime) || bEnd.Before(getReq.startTime) {
						return nil
					}

					return b.ForEach(func(k, v []byte) error {
						p, err := data.DecodePoint(v)
						if err != nil {
							log.Printf("badly encoded point %v", v)
							return nil
						}
						if p.Timestamp.Before(getReq.startTime) {
							return nil
						}
						if p.Timestamp.After(getReq.endTime) {
							return nil
						}

						points = append(points, p)
						return nil
					})
				})
			})
			getReq.ret <- points

		case <-tick.C:

			if err := db.Update(func(tx *bolt.Tx) error {
				return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
					if timestampFromBucket(name).Add(bdb.bucketSize + bdb.retention).Before(time.Now()) {
						return tx.DeleteBucket(name)
					}
					return nil
				})
			}); err != nil {
				log.Printf("db.Update failed: %v", err)
			}

		case <-ctx.Done():
			return nil
		}
	}

	return nil
}

// Push attempts to send data to DB.
func (bdb *boltDB) Push(points []data.Point) {
	bdb.pushCh <- points
}

func (bdb *boltDB) Get(startTime, endTime time.Time) map[string][]data.Point {
	req := getReq{
		startTime: startTime,
		endTime:   endTime,
		ret:       make(chan []data.Point),
	}
	bdb.getCh <- req
	return sortAndSplit(<-req.ret)
}

type boltDB struct {
	pushCh chan []data.Point
	getCh  chan getReq

	dbPath     string
	bucketSize time.Duration
	retention  time.Duration
}

type getReq struct {
	startTime time.Time
	endTime   time.Time
	ret       chan []data.Point
}

var _ db.Interface = (*boltDB)(nil)
var _ db.Source = (*boltDB)(nil)

func bucketFromTimestamp(timestamp time.Time, truncate time.Duration) []byte {
	ts := uint64(timestamp.Truncate(truncate).UnixNano())
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, ts)
	return b
}

func timestampFromBucket(b []byte) time.Time {
	if len(b) != 8 {
		return time.Time{}
	}
	return time.Unix(0, int64(binary.BigEndian.Uint64(b)))
}

func keyFromSeq(seq uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, seq)
	return b
}

func sortAndSplit(points []data.Point) map[string][]data.Point {
	ret := map[string][]data.Point{}

	for _, p := range points {
		ret[p.Address] = append(ret[p.Address], p)
	}

	for a := range ret {
		sort.Slice(ret[a], func(i, j int) bool { return ret[a][i].Timestamp.Before(ret[a][j].Timestamp) })
	}

	return ret
}
