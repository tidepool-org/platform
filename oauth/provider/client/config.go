package client

import (
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
)

type (
	ProviderConfig = oauthProvider.Config
	ClientConfig   = client.Config
)

type Config struct {
	*ProviderConfig
	ClientConfig *ClientConfig `json:"client,omitempty"`
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
		ProviderConfig: oauthProvider.NewConfig(),
		ClientConfig:   client.NewConfig(),
	}
}

func (c *Config) LoadFromConfigReporter(configReporter config.Reporter) error {
	if c.ProviderConfig != nil {
		if err := c.ProviderConfig.LoadFromConfigReporter(configReporter); err != nil {
			return err
		}
	}
	if c.ClientConfig != nil {
		if err := c.ClientConfig.LoadFromConfigReporter(configReporter.WithScopes("client")); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) Validate() error {
	if c.ProviderConfig == nil {
		return errors.New("provider config is missing")
	} else if err := c.ProviderConfig.Validate(); err != nil {
		return errors.Wrap(err, "provider config is invalid")
	}
	if c.ClientConfig == nil {
		return errors.New("client config is missing")
	} else if err := c.ClientConfig.Validate(); err != nil {
		return errors.Wrap(err, "client config is invalid")
	}
	return nil
}
