package config

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

//go:embed example.yaml
var ExampleConfig string

func Path(flag string) string {
	cfgPath := flag
	if cfgPath == "" {
		if os.Geteuid() == 0 {
			cfgPath = "/usr/local/ruuvi2db/config.yaml"
		} else {
			cfgPath = fmt.Sprintf("%s/.ruuvi2db/config.yaml", os.Getenv("HOME"))
		}
	}
	return cfgPath
}

func CreateExample(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return fmt.Errorf("file %q already exists", path)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create %q directory: %v", dir, err)
	}

	if err := ioutil.WriteFile(path, []byte(ExampleConfig), 0644); err != nil {
		return fmt.Errorf("failed to create %q file: %v", path, err)
	}

	return nil
}

func Read(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %v", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("failed to process %q: %v", path, err)
	}
	return cfg, nil
}

type Config struct {
	General struct {
		LogRate               time.Duration `yaml:"log_rate"`
		LogUnknownDevices     bool          `yaml:"log_unknown_devices"`
		MaxDatapointStaleness time.Duration `yaml:"max_datapoint_staleness"`
	} `yaml:"general"`

	Devices struct {
		RuuviTag []struct {
			MAC       string `yaml:"mac"`
			HumanName string `yaml:"human_name"`
		} `yaml:"ruuvi_tag"`
	} `yaml:"devices"`

	HTTP struct {
		Enable bool   `yaml:"enable"`
		Listen string `yaml:"listen"`
	} `yaml:"http"`

	Database struct {
		Path            string        `yaml:"path"`
		RetentionWindow time.Duration `yaml:"retention_window"`
	} `yaml:"database"`

	Debug struct {
		DumpBinaryLogs bool `yaml:"dump_binary_logs"`
		DumpReadings   bool `yaml:"dump_readings"`
		HTTPHandlers   bool `yaml:"http_handlers"`
	} `yaml:"debug"`

	Bluetooth struct {
		HCIID           int64         `yaml:"hci_id"` // Deprecated.
		WatchdogTimeout time.Duration `yaml:"watchdog_timeout"`
	} `yaml:"bluetooth"`
}
