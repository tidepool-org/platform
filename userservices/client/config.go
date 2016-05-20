package client

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
	"net/url"

	"github.com/tidepool-org/platform/app"
)

type Config struct {
	Address            string `json:"address"`
	RequestTimeout     int    `json:"requestTimeout"`
	ServerTokenSecret  string `json:"serverTokenSecret"`
	ServerTokenTimeout int    `json:"serverTokenTimeout"`
}

func (c *Config) Validate() error {
	if c.Address == "" {
		return app.Error("client", "address is missing")
	} else if _, err := url.Parse(c.Address); err != nil {
		return app.Error("client", "address is invalid")
	}
	if c.RequestTimeout < 0 {
		return app.Error("client", "request timeout is invalid")
	}
	if c.ServerTokenSecret == "" {
		return app.Error("client", "server token secret is missing")
	}
	if c.ServerTokenTimeout < 0 {
		return app.Error("client", "server token timeout is invalid")
	}

	if c.RequestTimeout == 0 {
		c.RequestTimeout = 60
	}
	if c.ServerTokenTimeout == 0 {
		c.ServerTokenTimeout = 60
	}
	return nil
}
