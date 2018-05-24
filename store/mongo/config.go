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
	Addresses        []string      `json:"addresses"`
	TLS              bool          `json:"tls"`
	Database         string        `json:"database"`
	CollectionPrefix string        `json:"collectionPrefix"`
	Username         *string       `json:"-"`
	Password         *string       `json:"-"`
	Timeout          time.Duration `json:"timeout"`
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
	if tlsString, err := configReporter.Get("tls"); err == nil {
		var tls bool
		tls, err = strconv.ParseBool(tlsString)
		if err != nil {
			return errors.New("tls is invalid")
		}
		c.TLS = tls
	}
	c.Database = configReporter.GetWithDefault("database", "")
	c.CollectionPrefix = configReporter.GetWithDefault("collection_prefix", "")
	if username, err := configReporter.Get("username"); err == nil {
		c.Username = pointer.FromString(username)
	}
	if password, err := configReporter.Get("password"); err == nil {
		c.Password = pointer.FromString(password)
	}
	if timeoutString, err := configReporter.Get("timeout"); err == nil {
		var timeout int64
		timeout, err = strconv.ParseInt(timeoutString, 10, 0)
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
