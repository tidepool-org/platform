package client

import (
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
)

type Config struct {
	Provider *oauthProvider.Config `json:",inline"`
	Client   *client.Config        `json:"client,omitempty"`
}

func NewConfigWithConfigReporter(configReporter config.Reporter) (*Config, error) {
	config := NewConfig()
	if err := config.LoadFromConfigReporter(configReporter); err != nil {
		return nil, err
	}
	return config, nil
}

func NewConfig() *Config {
	return &Config{
		Provider: oauthProvider.NewConfig(),
		Client:   client.NewConfig(),
	}
}

func (c *Config) LoadFromConfigReporter(configReporter config.Reporter) error {
	if c.Provider != nil {
		if err := c.Provider.LoadFromConfigReporter(configReporter); err != nil {
			return err
		}
	}
	if c.Client != nil {
		if err := c.Client.LoadFromConfigReporter(configReporter.WithScopes("client")); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) Validate() error {
	if c.Provider == nil {
		return errors.New("provider is missing")
	} else if err := c.Provider.Validate(); err != nil {
		return errors.Wrap(err, "provider is invalid")
	}
	if c.Client == nil {
		return errors.New("client is missing")
	} else if err := c.Client.Validate(); err != nil {
		return errors.Wrap(err, "client is invalid")
	}
	return nil
}
