package file

import (
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
)

type Config struct {
	Directory string
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Load(configReporter config.Reporter) error {
	if configReporter == nil {
		return errors.New("config reporter is missing")
	}

	c.Directory = configReporter.GetWithDefault("directory", c.Directory)
	return nil
}

func (c *Config) Validate() error {
	if c.Directory == "" {
		return errors.New("directory is missing")
	}
	return nil
}
