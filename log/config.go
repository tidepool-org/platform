package log

import (
	"github.com/sirupsen/logrus"

	"github.com/tidepool-org/platform/errors"
)

type Config struct {
	Level string `json:"level" default:"warn"`
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
