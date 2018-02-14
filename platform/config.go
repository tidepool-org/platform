package platform

import (
	"strconv"
	"time"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
)

type Config struct {
	*client.Config
	Timeout       time.Duration
	ServiceSecret string
}

func NewConfig() *Config {
	return &Config{
		Config:  client.NewConfig(),
		Timeout: 60 * time.Second,
	}
}

func (c *Config) Load(configReporter config.Reporter) error {
	if err := c.Config.Load(configReporter); err != nil {
		return err
	}

	if timeoutString, err := configReporter.Get("timeout"); err == nil {
		var timeout int64
		timeout, err = strconv.ParseInt(timeoutString, 10, 0)
		if err != nil {
			return errors.New("timeout is invalid")
		}
		c.Timeout = time.Duration(timeout) * time.Second
	}
	c.ServiceSecret = configReporter.GetWithDefault("service_secret", "")

	return nil
}

func (c *Config) Validate() error {
	if err := c.Config.Validate(); err != nil {
		return err
	}

	if c.Timeout <= 0 {
		return errors.New("timeout is invalid")
	}
	// TODO: Use once all services support service secret
	// if c.ServiceSecret == "" {
	// 	return errors.New("service secret is missing")
	// }

	return nil
}
