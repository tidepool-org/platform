package client

import (
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

type Config struct {
	*ExternalConfig
}

func NewConfig() *Config {
	return &Config{
		ExternalConfig: NewExternalConfig(),
	}
}

func (c *Config) Load(configReporter config.Reporter) error {
	return c.ExternalConfig.Load(configReporter.WithScopes("external"))
}

func (c *Config) Validate() error {
	return c.ExternalConfig.Validate()
}

type Client struct {
	*External
}

func NewClient(cfg *Config, name string, lgr log.Logger) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	}
	if name == "" {
		return nil, errors.New("name is missing")
	}
	if lgr == nil {
		return nil, errors.New("logger is missing")
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	extrnl, err := NewExternal(cfg.ExternalConfig, name, lgr)
	if err != nil {
		return nil, err
	}

	return &Client{
		External: extrnl,
	}, nil
}
