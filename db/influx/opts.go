package influx

type runOption func(*influxDB) error

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
