package log

import (
	"github.com/sirupsen/logrus"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
)

type Config struct {
	Level string
}

func NewConfig() *Config {
	return &Config{
		Level: "warn",
	}
}

func (c *Config) Load(configReporter config.Reporter) error {
	if configReporter == nil {
		return errors.New("log", "config reporter is missing")
	}

	c.Level = configReporter.StringOrDefault("level", "warn")

	return nil
}

func (c *Config) Validate() error {
	if level, err := logrus.ParseLevel(c.Level); err != nil {
		return errors.Wrap(err, "log", "level is invalid")
	} else if level == logrus.PanicLevel {
		return errors.New("log", "level is invalid")
	}

	return nil
}

func (c *Config) Clone() *Config {
	return &Config{
		Level: c.Level,
	}
}
