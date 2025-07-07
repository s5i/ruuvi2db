package reader

import (
	"time"
)

type Config struct {
	ProvidedEndpoints struct {
		Data string `yaml:"data"`
	} `yaml:"provided_endpoints"`

	Bluetooth struct {
		WatchdogTimeout time.Duration `yaml:"watchdog_timeout"`
	} `yaml:"bluetooth"`

	Data struct {
		MaxStaleness time.Duration `yaml:"max_staleness"`
		MACFilter    []string      `yaml:"mac_filter"`
	} `yaml:"data"`
}

func (cfg *Config) Sanitize() error {
	if cfg == nil {
		return nil
	}
	return nil
}
