package server

import (
	"os"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
)

type Config struct {
	Address            string
	TLS                bool
	TLSCertificateFile string
	TLSKeyFile         string
	Timeout            time.Duration
}

func NewConfig() *Config {
	return &Config{
		TLS:     true,
		Timeout: 60 * time.Second,
	}
}

func (c *Config) Load(configReporter config.Reporter) error {
	if configReporter == nil {
		return errors.New("config reporter is missing")
	}

	c.Address = configReporter.GetWithDefault("address", "")
	if tlsString, found := configReporter.Get("tls"); found {
		tls, err := strconv.ParseBool(tlsString)
		if err != nil {
			return errors.New("tls is invalid")
		}
		c.TLS = tls
	}
	c.TLSCertificateFile = configReporter.GetWithDefault("tls_certificate_file", "")
	c.TLSKeyFile = configReporter.GetWithDefault("tls_key_file", "")
	if timeoutString, found := configReporter.Get("timeout"); found {
		timeout, err := strconv.ParseInt(timeoutString, 10, 0)
		if err != nil {
			return errors.New("timeout is invalid")
		}
		c.Timeout = time.Duration(timeout) * time.Second
	}

	return nil
}

func (c *Config) Validate() error {
	if c.Address == "" {
		return errors.New("address is missing")
	}
	if c.TLS {
		if c.TLSCertificateFile == "" {
			return errors.New("tls certificate file is missing")
		} else if fileInfo, err := os.Stat(c.TLSCertificateFile); err != nil {
			if !os.IsNotExist(err) {
				return errors.Wrap(err, "unable to stat tls certificate file")
			}
			return errors.New("tls certificate file does not exist")
		} else if fileInfo.IsDir() {
			return errors.New("tls certificate file is a directory")
		}
		if c.TLSKeyFile == "" {
			return errors.New("tls key file is missing")
		} else if fileInfo, err := os.Stat(c.TLSKeyFile); err != nil {
			if !os.IsNotExist(err) {
				return errors.Wrap(err, "unable to stat tls key file")
			}
			return errors.New("tls key file does not exist")
		} else if fileInfo.IsDir() {
			return errors.New("tls key file is a directory")
		}
	}
	if c.Timeout <= 0 {
		return errors.New("timeout is invalid")
	}

	return nil
}
