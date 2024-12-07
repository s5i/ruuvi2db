package bolt

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strconv"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
	"github.com/s5i/ruuvi2db/data"
)

// Config contains options for Bolt database.
type Config struct {
	Path            string
	RetentionWindow time.Duration
	AllowRewrite    bool
}

// New returns an object that can be used to connect and push to Bolt DB.
func New() *DB {
	return &DB{
		pushPointsCh:    make(chan pushPointsReq),
		pointsCh:        make(chan pointsReq),
		setAliasCh:      make(chan setAliasReq),
		getAliasCh:      make(chan getAliasReq),
		retentionTicker: make(chan time.Time),
	}
}

// Run starts a connection to DB and handles Push calls.
func (d *DB) Run(ctx context.Context, cfg *Config) error {
	db, err := bolt.Open(cfg.Path, 0644, &bolt.Options{
		Timeout:   time.Second,
		MmapFlags: syscall.MAP_POPULATE,
	})
	if err != nil {
		return err
	}
	defer db.Close()

	if err := initDB(db, cfg.AllowRewrite); err != nil {
		return err
	}

	if cfg.RetentionWindow > 0 {
		period := cfg.RetentionWindow / 10
		if period > time.Hour {
			period = time.Hour
		}

		tick := time.NewTicker(period)
		defer tick.Stop()

		d.retentionTicker = tick.C
	}

	for {
		select {
		case req := <-d.pushPointsCh:
			req.execute(db)

		case req := <-d.pointsCh:
			req.execute(db)

		case req := <-d.setAliasCh:
			req.execute(db)

		case req := <-d.getAliasCh:
			req.execute(db)

		case <-d.retentionTicker:
			executeRetention(db, cfg.RetentionWindow)

		case <-ctx.Done():
			return nil
		}
	}
}

// PushPoints pushes data points to DB.
func (d *DB) PushPoints(points []*data.Point) error {
	respCh := make(chan pushPointsResp, 1)
	d.pushPointsCh <- pushPointsReq{
		points: points,
		respCh: respCh,
	}
	resp := <-respCh
	return resp.err
}

// Points returns data points between (startTime, endTime].
func (d *DB) Points(startTime, endTime time.Time) ([]*data.Point, error) {
	respCh := make(chan pointsResp, 1)
	d.pointsCh <- pointsReq{
		start:  startTime,
		end:    endTime,
		respCh: respCh,
	}
	resp := <-respCh
	return resp.points, resp.err
}

// SetAlias sets an alias for a MAC address.
func (d *DB) SetAlias(addr, name string) error {
	respCh := make(chan setAliasResp, 1)
	d.setAliasCh <- setAliasReq{
		addr:   addr,
		name:   name,
		respCh: respCh,
	}
	resp := <-respCh
	return resp.err
}

// Alias returns an alias for a MAC address.
func (d *DB) Alias(addr string) (string, error) {
	respCh := make(chan getAliasResp, 1)
	d.getAliasCh <- getAliasReq{
		addr:   addr,
		respCh: respCh,
	}
	resp := <-respCh
	return resp.alias, resp.err
}

type DB struct {
	pushPointsCh    chan pushPointsReq
	pointsCh        chan pointsReq
	setAliasCh      chan setAliasReq
	getAliasCh      chan getAliasReq
	retentionTicker <-chan time.Time
}

type pointsReq struct {
	start time.Time
	end   time.Time

	respCh chan pointsResp
}

type pointsResp struct {
	points    []*data.Point
	pointErrs []string
	err       error
}

func (req *pointsReq) execute(db *bolt.DB) {
	var points []*data.Point
	var pointErrs []string

	rStart, rEnd := req.start, req.end

	if err := db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte(pointsRoot))
		if root == nil {
			return nil
		}

		return root.ForEach(func(windowKey, _ []byte) error {
			windowBucket := root.Bucket(windowKey)
			if windowBucket == nil {
				return nil
			}

			wStart, wEnd := windowFromKey(windowKey)
			if !wStart.Before(rEnd) || !rStart.Before(wEnd) {
				return nil
			}

			return windowBucket.ForEach(func(addrKey, _ []byte) error {
				addrBucket := windowBucket.Bucket(addrKey)
				if addrBucket == nil {
					return nil
				}

				return addrBucket.ForEach(func(tsKey, dpRaw []byte) error {
					dp, err := data.DecodePoint(dpRaw)
					if err != nil {
						pointErrs = append(pointErrs, fmt.Sprintf("bad point @ %s / %X (%v) / %X (%v) / %X (%v): %X", pointsRoot, windowKey, wEnd, addrKey, net.HardwareAddr(addrKey), tsKey, tsFromKey(tsKey), dpRaw))
						return nil
					}

					ts := dp.Timestamp
					if !ts.After(rStart) || ts.After(rEnd) {
						return nil
					}

					points = append(points, dp)
					return nil
				})
			})
		})
	}); err != nil {
		req.respCh <- pointsResp{err: err}
	}
	req.respCh <- pointsResp{
		points:    points,
		pointErrs: pointErrs,
	}
}

type pushPointsReq struct {
	points []*data.Point

	respCh chan pushPointsResp
}

type pushPointsResp struct {
	err error
}

func (req *pushPointsReq) execute(db *bolt.DB) {
	if err := db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte(pointsRoot))
		if err != nil {
			return err
		}

		for _, dp := range req.points {
			dpRaw, err := dp.Encode()
			if err != nil {
				return err
			}

			windowKey, addrKey, tsKey, err := dpKeys(dp)
			if err != nil {
				return err
			}

			windowB, err := root.CreateBucketIfNotExists(windowKey)
			if err != nil {
				return err
			}

			addrB, err := windowB.CreateBucketIfNotExists(addrKey)
			if err != nil {
				return err
			}

			if err := addrB.Put(tsKey, dpRaw); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		req.respCh <- pushPointsResp{err: err}
	}
	req.respCh <- pushPointsResp{}
}

type setAliasReq struct {
	addr string
	name string

	respCh chan setAliasResp
}

type setAliasResp struct {
	err error
}

func (req *setAliasReq) execute(db *bolt.DB) {
	if err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(aliasesRoot))
		if err != nil {
			return err
		}

		if req.name == "" {
			return b.Delete([]byte(req.addr))
		}

		return b.Put([]byte(req.addr), []byte(req.name))
	}); err != nil {
		req.respCh <- setAliasResp{err: err}
	}

	req.respCh <- setAliasResp{}
}

type getAliasReq struct {
	addr string

	respCh chan getAliasResp
}

type getAliasResp struct {
	alias string
	err   error
}

func (req *getAliasReq) execute(db *bolt.DB) {
	var alias string
	if err := db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte(aliasesRoot))
		if root == nil {
			return nil
		}

		alias = string(root.Get([]byte(req.addr)))
		return nil
	}); err != nil {
		req.respCh <- getAliasResp{err: err}
	}
	req.respCh <- getAliasResp{alias: alias}
}

func executeRetention(db *bolt.DB, retention time.Duration) {
	if retention <= 0 {
		return
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte(pointsRoot))
		if root == nil {
			return nil
		}

		var toDelete [][]byte
		if err := root.ForEach(func(windowKey, _ []byte) error {
			if _, wEnd := windowFromKey(windowKey); wEnd.Add(retention).Before(time.Now()) {
				toDelete = append(toDelete, windowKey)
			}
			return nil
		}); err != nil {
			return err
		}

		for _, windowKey := range toDelete {
			if err := root.DeleteBucket(windowKey); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		log.Print(err)
	}
}

func windowKey(ts time.Time) []byte {
	_, r := windowFromTs(ts)
	return timestampKey(r)
}

func timestampKey(ts time.Time) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(int64(^uint64(0)>>1)-ts.UnixNano()))
	return b
}

func addrKey(addr string) ([]byte, error) {
	mac, err := net.ParseMAC(addr)
	if err != nil {
		return nil, err
	}
	return mac, nil
}

func dpKeys(dp *data.Point) (window []byte, addr []byte, ts []byte, err error) {
	addr, err = addrKey(dp.Address)
	if err != nil {
		return nil, nil, nil, err
	}

	window = windowKey(dp.Timestamp)
	ts = timestampKey(dp.Timestamp)

	return
}

func tsFromKey(b []byte) time.Time {
	return time.Unix(0, -int64(binary.LittleEndian.Uint64(b)-(^uint64(0)>>1)))
}

func windowFromTs(ts time.Time) (l time.Time, r time.Time) {
	tr := ts.Truncate(pointsWindowSize)
	if ts.Equal(tr) {
		return tr.Add(-pointsWindowSize), tr
	}
	return tr, tr.Add(pointsWindowSize)
}

func windowFromKey(b []byte) (l time.Time, r time.Time) {
	return windowFromTs(tsFromKey(b))
}

const (
	metadataRoot = `metadata`
	pointsRoot   = `points`
	aliasesRoot  = `aliases`

	metadataVersionKey     = `version`
	metadataVersionCurrent = 2

	pointsWindowSize = 24 * time.Hour
)

func initDB(db *bolt.DB, allowRewrite bool) error {
	v := 0
	if err := db.View(func(tx *bolt.Tx) error {
		if root := tx.Bucket([]byte(metadataRoot)); root != nil {
			sv64, err := strconv.ParseInt(string(root.Get([]byte(metadataVersionKey))), 0, 64)
			if err != nil {
				return err
			}
			v = int(sv64)
			return nil
		}

		// If there exist any buckets and we didn't short-circuit before, assume a pre-metadata ("v1") database.
		return tx.ForEach(func(_ []byte, _ *bolt.Bucket) error {
			v = 1
			return nil
		})
	}); err != nil {
		return err
	}

	for {
		switch v {
		case metadataVersionCurrent:
			return nil

		case 0:
			return db.Update(func(tx *bolt.Tx) error {
				if _, err := tx.CreateBucketIfNotExists([]byte(metadataRoot)); err != nil {
					return err
				}
				if _, err := tx.CreateBucketIfNotExists([]byte(pointsRoot)); err != nil {
					return err
				}
				if _, err := tx.CreateBucketIfNotExists([]byte(aliasesRoot)); err != nil {
					return err
				}
				return tx.Bucket([]byte(metadataRoot)).Put([]byte(metadataVersionKey), []byte(fmt.Sprint(metadataVersionCurrent)))
			})

		default:
			if !allowRewrite {
				return fmt.Errorf("detected old DB schema version %d, want %d; requires AllowRewrite to proceed", v, metadataVersionCurrent)
			}
		}

		switch v {
		case 1:
			if err := rewriteV1toV2(db); err != nil {
				return err
			}
			v = 2
		}
	}
}

func rewriteV1toV2(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(metadataRoot)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(pointsRoot)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(aliasesRoot)); err != nil {
			return err
		}
		if err := tx.Bucket([]byte(metadataRoot)).Put([]byte(metadataVersionKey), []byte(fmt.Sprint(2))); err != nil {
			return err
		}

		root := tx.Bucket([]byte(pointsRoot))

		var toDelete [][]byte

		if err := tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			switch {
			case bytes.Equal(name, []byte(metadataRoot)):
				return nil
			case bytes.Equal(name, []byte(pointsRoot)):
				return nil
			case bytes.Equal(name, []byte(aliasesRoot)):
				return nil
			}

			toDelete = append(toDelete, bytes.Clone(name))

			return b.ForEach(func(_, v []byte) error {
				dp, err := data.DecodePoint(v)
				if err != nil {
					return err
				}

				windowKey, addrKey, tsKey, err := dpKeys(dp)
				if err != nil {
					return err
				}

				windowB, err := root.CreateBucketIfNotExists(windowKey)
				if err != nil {
					return err
				}

				addrB, err := windowB.CreateBucketIfNotExists(addrKey)
				if err != nil {
					return err
				}

				if err := addrB.Put(tsKey, v); err != nil {
					return err
				}
				return nil
			})
		}); err != nil {
			return err
		}

		for _, name := range toDelete {
			if err := tx.DeleteBucket(name); err != nil {
				return err
			}
		}

		return nil
	})
}
