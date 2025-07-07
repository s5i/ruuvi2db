package storage

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	ProvidedEndpoints struct {
		Data  string `yaml:"data"`
		Admin string `yaml:"admin"`
	} `yaml:"provided_endpoints"`

	ConsumedEndpoints struct {
		Reader string `yaml:"reader"`
	} `yaml:"consumed_endpoints"`

	ReaderConsumer struct {
		QueryPeriod  time.Duration `yaml:"query_period"`
		MaxStaleness time.Duration `yaml:"max_staleness"`
		MACFilter    []string      `yaml:"mac_filter"`
	} `yaml:"reader_consumer"`

	Database struct {
		Bolt struct {
			Path              string        `yaml:"path"`
			RetentionWindow   time.Duration `yaml:"retention_window"`
			AllowSchemaUpdate bool          `yaml:"allow_schema_update"`
		} `yaml:"bolt"`
	} `yaml:"database"`
}

func (cfg *Config) Sanitize() error {
	if cfg == nil {
		return nil
	}

	cfg.Database.Bolt.Path = sanitizePath(cfg.Database.Bolt.Path)

	return nil
}

func sanitizePath(path string) string {
	if x, ok := strings.CutPrefix(path, "~/"); ok {
		return filepath.Join(os.Getenv("HOME"), x)
	}

	return path
}
