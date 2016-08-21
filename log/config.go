package log

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"github.com/Sirupsen/logrus"

	"github.com/tidepool-org/platform/app"
)

type Config struct {
	Level string `json:"level" default:"warn"`
}

func (c *Config) Validate() error {
	if level, err := logrus.ParseLevel(c.Level); err != nil {
		return app.ExtError(err, "log", "level is invalid")
	} else if level == logrus.PanicLevel {
		return app.Error("log", "level is invalid")
	}
	return nil
}

func (c *Config) Clone() *Config {
	return &Config{
		Level: c.Level,
	}
}
