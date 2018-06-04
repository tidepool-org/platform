package mongo

import (
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type Config struct {
	*storeStructuredMongo.Config
	PasswordSalt string
}

func NewConfig() *Config {
	return &Config{
		Config: storeStructuredMongo.NewConfig(),
	}
}

func (c *Config) Load(configReporter config.Reporter) error {
	if c.Config == nil {
		return errors.New("config is missing")
	}
	if err := c.Config.Load(configReporter); err != nil {
		return err
	}

	c.PasswordSalt = configReporter.GetWithDefault("password_salt", "")

	return nil
}

func (c *Config) Validate() error {
	if c.Config == nil {
		return errors.New("config is missing")
	}
	if err := c.Config.Validate(); err != nil {
		return err
	}

	if c.PasswordSalt == "" {
		return errors.New("password salt is missing")
	}

	return nil
}
