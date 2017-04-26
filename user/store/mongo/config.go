package mongo

import (
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/store/mongo"
)

type Config struct {
	*mongo.Config `anonymous:"true"`
	PasswordSalt  string `json:"passwordSalt"`
}

func (c *Config) Validate() error {
	if c.Config == nil {
		return errors.New("mongo", "config is missing")
	}
	if err := c.Config.Validate(); err != nil {
		return err
	}
	if c.PasswordSalt == "" {
		return errors.New("mongo", "password salt is missing")
	}
	return nil
}

func (c *Config) Clone() *Config {
	clone := &Config{
		Config:       c.Config.Clone(),
		PasswordSalt: c.PasswordSalt,
	}
	return clone
}
