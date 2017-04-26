package client

import (
	"net/url"

	"github.com/tidepool-org/platform/app"
)

type Config struct {
	Address        string `json:"address"`
	RequestTimeout int    `json:"requestTimeout"`
}

func (c *Config) Validate() error {
	if c.Address == "" {
		return app.Error("client", "address is missing")
	} else if _, err := url.Parse(c.Address); err != nil {
		return app.Error("client", "address is invalid")
	}
	if c.RequestTimeout < 0 {
		return app.Error("client", "request timeout is invalid")
	}

	if c.RequestTimeout == 0 {
		c.RequestTimeout = 60
	}
	return nil
}

func (c *Config) Clone() *Config {
	return &Config{
		Address:        c.Address,
		RequestTimeout: c.RequestTimeout,
	}
}
