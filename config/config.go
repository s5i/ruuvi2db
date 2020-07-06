package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	proto "github.com/golang/protobuf/proto"
)

func Path(flag string) string {
	cfgPath := flag
	if cfgPath == "" {
		if os.Geteuid() == 0 {
			cfgPath = "/usr/local/ruuvi2db/config.textproto"
		} else {
			cfgPath = fmt.Sprintf("%s/.ruuvi2db/config.textproto", os.Getenv("HOME"))
		}
	}
	return cfgPath
}

func CreateExample(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return fmt.Errorf("file %s already exists", path)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create %s directory: %v", err)
	}

	if err := ioutil.WriteFile(path, []byte(ExampleConfig), 0644); err != nil {
		return fmt.Errorf("failed to create %s file: %v", path, err)
	}

	return nil
}

func Read(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %v", path, err)
	}

	cfg := &Config{}
	if err := proto.UnmarshalText(string(b), cfg); err != nil {
		return nil, fmt.Errorf("failed to process %s: %v", path, err)
	}
	return cfg, nil
}
