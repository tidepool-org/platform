package mongo

import (
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/store/mongo"
)

type Config struct {
	*mongo.Config `anonymous:"true"`
	Secret        string `json:"secret"`
}

func (c *Config) Validate() error {
	if c.Config == nil {
		return errors.New("mongo", "config is missing")
	}
	if err := c.Config.Validate(); err != nil {
		return err
	}
	if c.Secret == "" {
		return errors.New("mongo", "secret is missing")
	}
	return nil
}

func (c *Config) Clone() *Config {
	clone := &Config{
		Config: c.Config.Clone(),
		Secret: c.Secret,
	}
	return clone
}
