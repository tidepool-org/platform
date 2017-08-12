package client

import (
	"strconv"
	"time"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
)

type Config struct {
	*client.Config
	ServerTokenSecret  string
	ServerTokenTimeout time.Duration
}

func NewConfig() *Config {
	return &Config{
		Config:             client.NewConfig(),
		ServerTokenTimeout: 3600 * time.Second,
	}
}

func (c *Config) Load(configReporter config.Reporter) error {
	if err := c.Config.Load(configReporter); err != nil {
		return err
	}

	c.ServerTokenSecret = configReporter.GetWithDefault("server_token_secret", "")
	if serverTokenTimeoutString, found := configReporter.Get("server_token_timeout"); found {
		serverTokenTimeout, err := strconv.ParseInt(serverTokenTimeoutString, 10, 0)
		if err != nil {
			return errors.New("client", "server token timeout is invalid")
		}
		c.ServerTokenTimeout = time.Duration(serverTokenTimeout) * time.Second
	}

	return nil
}

func (c *Config) Validate() error {
	if err := c.Config.Validate(); err != nil {
		return err
	}

	if c.ServerTokenSecret == "" {
		return errors.New("client", "server token secret is missing")
	}
	if c.ServerTokenTimeout <= 0 {
		return errors.New("client", "server token timeout is invalid")
	}

	return nil
}
