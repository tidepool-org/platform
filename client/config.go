package client

import (
	"net/url"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
)

type Config struct {
	Address   string
	UserAgent string
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Load(configReporter config.Reporter) error {
	if configReporter == nil {
		return errors.New("config reporter is missing")
	}

	c.Address = configReporter.GetWithDefault("address", "")
	if userAgent, found := configReporter.Get("user_agent"); found {
		c.UserAgent = userAgent
	}

	return nil
}

func (c *Config) Validate() error {
	if c.Address == "" {
		return errors.New("address is missing")
	}
	if _, err := url.Parse(c.Address); err != nil {
		return errors.New("address is invalid")
	}
	if c.UserAgent == "" {
		return errors.New("user agent is missing")
	}

	return nil
}
