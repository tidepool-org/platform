package mongo

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
)

type Config struct {
	Addresses  []string
	TLS        bool
	Database   string
	Collection string
	Username   *string
	Password   *string
	Timeout    time.Duration
}

func NewConfig() *Config {
	return &Config{
		TLS:     true,
		Timeout: 60 * time.Second,
	}
}

func (c *Config) Load(configReporter config.Reporter) error {
	if configReporter == nil {
		return errors.New("mongo", "config reporter is missing")
	}

	c.Addresses = SplitAddresses(configReporter.StringOrDefault("addresses", ""))
	if tlsString, found := configReporter.String("tls"); found {
		tls, err := strconv.ParseBool(tlsString)
		if err != nil {
			return errors.New("mongo", "tls is invalid")
		}
		c.TLS = tls
	}
	c.Database = configReporter.StringOrDefault("database", "")
	c.Collection = configReporter.StringOrDefault("collection", "")
	if username, found := configReporter.String("username"); found {
		c.Username = pointer.String(username)
	}
	if password, found := configReporter.String("password"); found {
		c.Password = pointer.String(password)
	}
	if timeoutString, found := configReporter.String("timeout"); found {
		timeout, err := strconv.ParseInt(timeoutString, 10, 0)
		if err != nil {
			return errors.New("mongo", "timeout is invalid")
		}
		c.Timeout = time.Duration(timeout) * time.Second
	}

	return nil
}

func (c *Config) Validate() error {
	if len(c.Addresses) < 1 {
		return errors.New("mongo", "addresses is missing")
	}
	for _, address := range c.Addresses {
		if address == "" {
			return errors.New("mongo", "address is missing")
		}
		if _, err := url.Parse(address); err != nil {
			return errors.New("mongo", "address is invalid")
		}
	}
	if c.Database == "" {
		return errors.New("mongo", "database is missing")
	}
	if c.Collection == "" {
		return errors.New("mongo", "collection is missing")
	}
	if c.Timeout <= 0 {
		return errors.New("mongo", "timeout is invalid")
	}

	return nil
}

func (c *Config) Clone() *Config {
	clone := &Config{
		TLS:        c.TLS,
		Database:   c.Database,
		Collection: c.Collection,
		Timeout:    c.Timeout,
	}
	if c.Addresses != nil {
		clone.Addresses = append([]string{}, c.Addresses...)
	}
	if c.Username != nil {
		clone.Username = pointer.String(*c.Username)
	}
	if c.Password != nil {
		clone.Password = pointer.String(*c.Password)
	}

	return clone
}

func SplitAddresses(addressesString string) []string {
	addressses := []string{}
	for _, address := range strings.Split(addressesString, ",") {
		address = strings.TrimSpace(address)
		if address != "" {
			addressses = append(addressses, address)
		}
	}
	return addressses
}
