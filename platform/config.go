package platform

import (
	"github.com/kelseyhightower/envconfig"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/config"
)

// Config extends client.Config with additional properties.
type Config struct {
	*client.Config
	ServiceSecret string `envconfig:"TIDEPOOL_SERVICE_SECRET"` // this should be overridden for loaders using envconfig
}

func NewConfig() *Config {
	return &Config{
		Config: client.NewConfig(),
	}
}

func (c *Config) Load(loader ConfigLoader) error {
	return loader.Load(c)
}

// ConfigLoader abstracts the method by which config values are loaded.
type ConfigLoader interface {
	Load(*Config) error
}

// configReporterLoader adapts config.Reporter to implement ConfigLoader.
type configReporterLoader struct {
	Reporter config.Reporter
	client.ConfigLoader
}

func NewConfigReporterLoader(reporter config.Reporter) *configReporterLoader {
	return &configReporterLoader{
		ConfigLoader: client.NewConfigReporterLoader(reporter),
		Reporter:     reporter,
	}
}

// LoadPlatform implements ConfigLoader.
func (l *configReporterLoader) Load(cfg *Config) error {
	if err := l.ConfigLoader.Load(cfg.Config); err != nil {
		return err
	}
	cfg.ServiceSecret = l.Reporter.GetWithDefault("service_secret", cfg.ServiceSecret)
	return nil
}

// envconfigLoader adapts envconfig to implement ConfigLoader.
type envconfigLoader struct {
	client.ConfigLoader
}

// NewEnvconfigLoader loads values via envconfig.
//
// If loader is nil, it defaults to envconfig for client values.
func NewEnvconfigLoader(loader client.ConfigLoader) *envconfigLoader {
	if loader == nil {
		loader = client.NewEnvconfigLoader()
	}
	return &envconfigLoader{
		ConfigLoader: loader,
	}
}

// Load implements ConfigLoader.
func (l *envconfigLoader) Load(cfg *Config) error {
	if err := l.ConfigLoader.Load(cfg.Config); err != nil {
		return err
	}
	return envconfig.Process(client.EnvconfigEmptyPrefix, cfg)
}
