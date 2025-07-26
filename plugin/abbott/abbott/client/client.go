package client

import (
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/oauth"
)

type Config struct {
	*client.Config
}

func NewConfig() *Config {
	return &Config{
		Config: client.NewConfig(),
	}
}

func (c *Config) LoadFromConfigReporter(reporter config.Reporter) error {
	return nil
}

func (c *Config) Validate() error {
	return nil
}

type ClientDependencies struct {
	Config            *Config
	TokenSourceSource oauth.TokenSourceSource
}

type Client struct{}

func NewClient(clientDependencies ClientDependencies) (*Client, error) {
	return &Client{}, nil
}
