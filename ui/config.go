package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ProvidedEndpoints struct {
		UI string `yaml:"ui"`
	} `yaml:"provided_endpoints"`

	ConsumedEndpoints struct {
		Storage string `yaml:"storage"`
	} `yaml:"consumed_endpoints"`
}

func (cfg *Config) Sanitize() error {
	if cfg == nil {
		return nil
	}
	return nil
}

func ReadConfig(path string) (*Config, error) {
	switch {
	case path != "":
		path = sanitizePath(path)
	case os.Geteuid() == 0:
		path = "/usr/local/ruuvi2db/ui.cfg"
	default:
		path = fmt.Sprintf("%s/.ruuvi2db/ui.cfg", os.Getenv("HOME"))
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
