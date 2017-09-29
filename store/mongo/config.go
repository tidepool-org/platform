package mongo

import (
	"net/url"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
)

type Config struct {
	Addresses        []string
	TLS              bool
	Database         string
	CollectionPrefix string
	Username         *string
	Password         *string
	Timeout          time.Duration
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

	c.Addresses = SplitAddresses(configReporter.GetWithDefault("addresses", ""))
	if tlsString, found := configReporter.Get("tls"); found {
		tls, err := strconv.ParseBool(tlsString)
		if err != nil {
			return errors.New("tls is invalid")
		}
		c.TLS = tls
	}
	c.Database = configReporter.GetWithDefault("database", "")
	c.CollectionPrefix = configReporter.GetWithDefault("collection_prefix", "")
	if username, found := configReporter.Get("username"); found {
		c.Username = pointer.String(username)
	}
	if password, found := configReporter.Get("password"); found {
		c.Password = pointer.String(password)
	}
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
	if len(c.Addresses) == 0 {
		return errors.New("addresses is missing")
	}
	for _, address := range c.Addresses {
		if address == "" {
			return errors.New("address is missing")
		}
		if _, err := url.Parse(address); err != nil {
			return errors.New("address is invalid")
		}
	}
	if c.Database == "" {
		return errors.New("database is missing")
	}
	if c.Timeout <= 0 {
		return errors.New("timeout is invalid")
	}

	return nil
}

func SplitAddresses(addresses string) []string {
	return config.SplitTrimCompact(addresses)
}
