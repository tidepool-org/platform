package server

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
	"os"

	"github.com/tidepool-org/platform/app"
)

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
		} else if fileInfo, err := os.Stat(c.TLS.CertificateFile); err != nil {
			if !os.IsNotExist(err) {
				return app.ExtError(err, "server", "unable to stat tls certificate file")
			}
			return app.Error("server", "tls certificate file does not exist")
		} else if fileInfo.IsDir() {
			return app.Error("server", "tls certificate file is a directory")
		}
		if c.TLS.KeyFile == "" {
			return app.Error("server", "tls key file is missing")
		} else if fileInfo, err := os.Stat(c.TLS.KeyFile); err != nil {
			if !os.IsNotExist(err) {
				return app.ExtError(err, "server", "unable to stat tls key file")
			}
			return app.Error("server", "tls key file does not exist")
		} else if fileInfo.IsDir() {
			return app.Error("server", "tls key file is a directory")
		}
	}
	if c.Timeout < 0 {
		return app.Error("server", "timeout is invalid")
	}
	return nil
}
