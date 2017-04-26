package server

import (
	"os"

	"github.com/tidepool-org/platform/errors"
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
		return errors.New("server", "address is missing")
	}
	if c.TLS != nil {
		if c.TLS.CertificateFile == "" {
			return errors.New("server", "tls certificate file is missing")
		} else if fileInfo, err := os.Stat(c.TLS.CertificateFile); err != nil {
			if !os.IsNotExist(err) {
				return errors.Wrap(err, "server", "unable to stat tls certificate file")
			}
			return errors.New("server", "tls certificate file does not exist")
		} else if fileInfo.IsDir() {
			return errors.New("server", "tls certificate file is a directory")
		}
		if c.TLS.KeyFile == "" {
			return errors.New("server", "tls key file is missing")
		} else if fileInfo, err := os.Stat(c.TLS.KeyFile); err != nil {
			if !os.IsNotExist(err) {
				return errors.Wrap(err, "server", "unable to stat tls key file")
			}
			return errors.New("server", "tls key file does not exist")
		} else if fileInfo.IsDir() {
			return errors.New("server", "tls key file is a directory")
		}
	}
	if c.Timeout < 0 {
		return errors.New("server", "timeout is invalid")
	}
	return nil
}

func (c *Config) Clone() *Config {
	clone := &Config{
		Address: c.Address,
		Timeout: c.Timeout,
	}
	if c.TLS != nil {
		clone.TLS = &TLS{
			CertificateFile: c.TLS.CertificateFile,
			KeyFile:         c.TLS.KeyFile,
		}
	}
	return clone
}
