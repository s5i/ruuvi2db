package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/s5i/ruuvi2db/reader"
	"github.com/s5i/ruuvi2db/storage"
	"github.com/s5i/ruuvi2db/ui"
	"gopkg.in/yaml.v2"

	_ "embed"
)

type Config struct {
	Reader  *reader.Config  `yaml:"reader"`
	Storage *storage.Config `yaml:"storage"`
	UI      *ui.Config      `yaml:"ui"`

	DryRun bool `yaml:"dry_run"`
}

func (cfg *Config) Sanitize() error {
	if err := cfg.Reader.Sanitize(); err != nil {
		return err
	}
	if err := cfg.Storage.Sanitize(); err != nil {
		return err
	}
	if err := cfg.UI.Sanitize(); err != nil {
		return err
	}
	return nil
}

func ReadConfig(path string) (*Config, error) {
	switch {
	case path != "":
		path = sanitizePath(path)
	case os.Geteuid() == 0:
		path = "/usr/local/ruuvi2db/ruuvi2db.cfg"
	default:
		path = fmt.Sprintf("%s/.ruuvi2db/ruuvi2db.cfg", os.Getenv("HOME"))
	}

	if _, rErr := os.Stat(path); os.IsNotExist(rErr) {
		if wErr := createExample(path); wErr != nil {
			return nil, fmt.Errorf("%v; %v", rErr, wErr)
		}
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %v", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %q: %v", path, err)
	}

	if cfg.DryRun {
		return nil, fmt.Errorf("%q configured with dry_run", path)
	}

	return cfg, nil
}

func sanitizePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(os.Getenv("HOME"), strings.TrimPrefix(path, "~/"))
	}
	return path
}

func createExample(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create initial config directory %q: %v", dir, err)
	}

	if err := os.WriteFile(path, []byte(example), 0644); err != nil {
		return fmt.Errorf("failed to create initial config file %q: %v", path, err)
	}

	return nil
}

//go:embed example.cfg
var example []byte
