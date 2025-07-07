package ui

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
