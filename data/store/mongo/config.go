package mongo

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"strings"
	"time"

	"github.com/tidepool-org/platform/app"
)

type Config struct {
	Addresses  string         `yaml:"addresses"` // TODO: This should be an array, but configor does not support that. Bleech! Fix?
	Database   string         `yaml:"database"`
	Collection string         `yaml:"collection"`
	Username   *string        `yaml:"username"`
	Password   *string        `yaml:"password"`
	Timeout    *time.Duration `yaml:"timeout"`
	SSL        bool           `yaml:"ssl"`
}

func (c *Config) Validate() error {
	addresses := strings.Split(c.Addresses, ",")
	if len(addresses) < 1 {
		return app.Error("mongo", "addresses is missing")
	}
	if c.Database == "" {
		return app.Error("mongo", "database is missing")
	}
	if c.Collection == "" {
		return app.Error("mongo", "collection is missing")
	}
	return nil
}
