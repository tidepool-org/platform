package client

import (
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/platform"
)

type Config struct {
	*platform.Config
	*ExternalConfig
}

func NewConfig() *Config {
	return &Config{
		Config:         platform.NewConfig(),
		ExternalConfig: NewExternalConfig(),
	}
}

func (c *Config) Load(configReporter config.Reporter) error {
	if err := c.Config.Load(configReporter); err != nil {
		return err
	}
	return c.ExternalConfig.Load(configReporter.WithScopes("external"))
}

func (c *Config) Validate() error {
	if err := c.Config.Validate(); err != nil {
		return err
	}
	return c.ExternalConfig.Validate()
}

type Client struct {
	client *platform.Client
	*External
}

func NewClient(cfg *Config, authorizeAs platform.AuthorizeAs, name string, lgr log.Logger) (*Client, error) {
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

	clnt, err := platform.NewClient(cfg.Config, authorizeAs)
	if err != nil {
		return nil, err
	}

	extrnl, err := NewExternal(cfg.ExternalConfig, authorizeAs, name, lgr)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:   clnt,
		External: extrnl,
	}, nil
}
