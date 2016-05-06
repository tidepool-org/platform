package server

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import "github.com/tidepool-org/platform/app"

type Config struct {
	Address string `json:"address"`
	TLS     *TLS   `json:"tls"`
	Timeout int    `json:"timeout" default:"60"`
}

type TLS struct {
	CertificateFile string `json:"certificateFile"`
	KeyFile         string `json:"keyFile"`
}

func (c *Config) Validate() error {
	if c.Address == "" {
		return app.Error("server", "address is missing")
	}
	if c.TLS != nil {
		if c.TLS.CertificateFile == "" {
			return app.Error("server", "tls certificate file is missing")
		}
		if c.TLS.KeyFile == "" {
			return app.Error("server", "tls key file is missing")
		}
	}
	if c.Timeout < 0 {
		return app.Error("server", "timeout is invalid")
	}
	return nil
}
