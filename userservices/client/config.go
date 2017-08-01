package client

import (
	"net/url"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
)

type Config struct {
	Address            string
	Timeout            time.Duration
	ServerTokenSecret  string
	ServerTokenTimeout time.Duration
}

func NewConfig() *Config {
	return &Config{
		Timeout:            60 * time.Second,
		ServerTokenTimeout: 3600 * time.Second,
	}
}

func (c *Config) Load(configReporter config.Reporter) error {
	if configReporter == nil {
		return errors.New("client", "config reporter is missing")
	}

	c.Address = configReporter.StringOrDefault("address", "")
	if timeoutString, found := configReporter.String("timeout"); found {
		timeout, err := strconv.ParseInt(timeoutString, 10, 0)
		if err != nil {
			return errors.New("client", "timeout is invalid")
		}
		c.Timeout = time.Duration(timeout) * time.Second
	}
	c.ServerTokenSecret = configReporter.StringOrDefault("server_token_secret", "")
	if serverTokenTimeoutString, found := configReporter.String("server_token_timeout"); found {
		serverTokenTimeout, err := strconv.ParseInt(serverTokenTimeoutString, 10, 0)
		if err != nil {
			return errors.New("client", "server token timeout is invalid")
		}
		c.ServerTokenTimeout = time.Duration(serverTokenTimeout) * time.Second
	}

	return nil
}

func (c *Config) Validate() error {
	if c.Address == "" {
		return errors.New("client", "address is missing")
	}
	if _, err := url.Parse(c.Address); err != nil {
		return errors.New("client", "address is invalid")
	}
	if c.Timeout <= 0 {
		return errors.New("client", "timeout is invalid")
	}
	if c.ServerTokenSecret == "" {
		return errors.New("client", "server token secret is missing")
	}
	if c.ServerTokenTimeout <= 0 {
		return errors.New("client", "server token timeout is invalid")
	}

	return nil
}

func (c *Config) Clone() *Config {
	return &Config{
		Address:            c.Address,
		Timeout:            c.Timeout,
		ServerTokenSecret:  c.ServerTokenSecret,
		ServerTokenTimeout: c.ServerTokenTimeout,
	}
}
