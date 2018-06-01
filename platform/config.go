package platform

import (
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/config"
)

type Config struct {
	*client.Config
	ServiceSecret string
}

func NewConfig() *Config {
	return &Config{
		Config: client.NewConfig(),
	}
}

func (c *Config) Load(configReporter config.Reporter) error {
	if err := c.Config.Load(configReporter); err != nil {
		return err
	}

	c.ServiceSecret = configReporter.GetWithDefault("service_secret", c.ServiceSecret)

	return nil
}
