package client

import (
	"net/url"

	"github.com/kelseyhightower/envconfig"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
)

type Config struct {
	Address   string // this should be overridden for loaders using envconfig
	UserAgent string `envconfig:"TIDEPOOL_USER_AGENT"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Load(loader ConfigLoader) error {
	return loader.Load(c)
}

func (c *Config) Validate() error {
	if c.Address == "" {
		return errors.New("address is missing")
	} else if _, err := url.Parse(c.Address); err != nil {
		return errors.New("address is invalid")
	}
	if c.UserAgent == "" {
		return errors.New("user agent is missing")
	}

	return nil
}

// ConfigLoader abstracts the method by which config values are loaded.
type ConfigLoader interface {
	// Load sets config values for the properties of Config.
	Load(*Config) error
}

// configReporterLoader adapts a config.Reporter to implement ConfigLoader.
type configReporterLoader struct {
	Reporter config.Reporter
}

func NewConfigReporterLoader(reporter config.Reporter) *configReporterLoader {
	return &configReporterLoader{
		Reporter: reporter,
	}
}

// Load implements ConfigLoader.
func (l *configReporterLoader) Load(cfg *Config) error {
	cfg.Address = l.Reporter.GetWithDefault("address", cfg.Address)
	cfg.UserAgent = l.Reporter.GetWithDefault("user_agent", cfg.UserAgent)
	return nil
}

// EnvconfigEmptyPrefix should be the empty string.
//
// By forcing the use of the environment variable name in each tag, we aim to
// make the code more easily searchable.
const EnvconfigEmptyPrefix = ""

// envconfigLoader adapts envconfig to implement ConfigLoader.
type envconfigLoader struct{}

func NewEnvconfigLoader() *envconfigLoader {
	return &envconfigLoader{}
}

// Load implements ConfigLoader.
func (l *envconfigLoader) Load(cfg *Config) error {
	return envconfig.Process(EnvconfigEmptyPrefix, cfg)
}
