package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
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

func ReadConfig(path string) (*Config, error) {
	switch {
	case path != "":
		path = sanitizePath(path)
	case os.Geteuid() == 0:
		path = "/usr/local/ruuvi2db/storage.cfg"
	default:
		path = fmt.Sprintf("%s/.ruuvi2db/storage.cfg", os.Getenv("HOME"))
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %v", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %q: %v", path, err)
	}

	return cfg, nil
}

func sanitizePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(os.Getenv("HOME"), strings.TrimPrefix(path, "~/"))
	}
	return path
}
