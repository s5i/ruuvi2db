package database

import (
	"bytes"
	"context"
	"encoding/binary"
	"log"
	"sort"
	"time"

	"github.com/boltdb/bolt"
	"github.com/s5i/ruuvi2db/config"
	"github.com/s5i/ruuvi2db/data"
)

// NewDB returns an object that can be used to connect and push to Bolt DB.
func NewDB() *boltDB {
	return &boltDB{
		pushCh:    make(chan []data.Point, 1),
		getCh:     make(chan getReq),
		retention: 7 * 24 * time.Hour,
	}
}

// Run starts a connection to DB and handles Push calls.
func (bdb *boltDB) Run(ctx context.Context, cfg *config.Config) error {
	bdb.dbPath = cfg.Database.Path

	db, err := bolt.Open(bdb.dbPath, 0644, &bolt.Options{Timeout: time.Second})
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
					b, err := tx.CreateBucketIfNotExists(bucketFromTimestamp(p.Timestamp, bucketSize))
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
					bEnd := bStart.Add(bucketSize)
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
					if timestampFromBucket(name).Add(bucketSize + bdb.retention).Before(time.Now()) {
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

// Rewrite rewrites the database.
// This can resolve issues such as points being incorrectly bucketed due to bucket size change.
func (bdb *boltDB) Rewrite() error {
	db, err := bolt.Open(bdb.dbPath, 0644, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		tmp := []byte("temp")
		newB, err := tx.CreateBucketIfNotExists(tmp)
		if err != nil {
			return err
		}
		if err := tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			if bytes.Equal(name, tmp) {
				return nil
			}

			if err := b.ForEach(func(_, v []byte) error {
				seq, err := newB.NextSequence()
				if err != nil {
					return err
				}

				k := keyFromSeq(seq)
				return newB.Put(k, v)
			}); err != nil {
				return err
			}
			return tx.DeleteBucket(name)
		}); err != nil {
			return err
		}
		if err := newB.ForEach(func(_, v []byte) error {
			p, err := data.DecodePoint(v)
			if err != nil {
				log.Printf("badly encoded point %v", v)
				return nil
			}

			b, err := tx.CreateBucketIfNotExists(bucketFromTimestamp(p.Timestamp, bucketSize))
			if err != nil {
				return err
			}

			seq, err := b.NextSequence()
			if err != nil {
				return err
			}
			k := keyFromSeq(seq)

			return b.Put(k, v)
		}); err != nil {
			return err
		}
		return tx.DeleteBucket(tmp)
	})
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

	dbPath    string
	retention time.Duration
}

type getReq struct {
	startTime time.Time
	endTime   time.Time
	ret       chan []data.Point
}

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

const bucketSize = 24 * time.Hour
