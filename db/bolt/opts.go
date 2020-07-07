package bolt

import "time"

type runOption func(*boltDB) error

// WithBucketSize sets the bucket size for underlying DB.
func WithBucketSize(bs time.Duration) runOption {
	return func(bdb *boltDB) error {
		bdb.bucketSize = bs
		return nil
	}
}

// WithRetention sets data retention horizon.
func WithRetention(bs time.Duration) runOption {
	return func(bdb *boltDB) error {
		bdb.retention = bs
		return nil
	}
}
