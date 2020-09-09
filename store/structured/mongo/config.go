package mongo

import (
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	mgo "github.com/globalsign/mgo"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
)

//Config describe parameters need to make a connection to a Mongo database
type Config struct {
	Scheme                 string `json:"scheme"`
	addresses              []string
	TLS                    bool                   `json:"tls"`
	Database               string                 `json:"database"`
	CollectionPrefix       string                 `json:"collectionPrefix"`
	Username               *string                `json:"-"`
	Password               *string                `json:"-"`
	Timeout                time.Duration          `json:"timeout"`
	OptParams              *string                `json:"optParams"`
	WaitConnectionInterval time.Duration          `json:"waitConnectionInterval"`
	MaxConnectionAttempts  int64                  `json:"maxConnectionAttempts"`
	Indexes                map[string][]mgo.Index `json:"indexes"`
	adressMux              sync.Mutex
}

//NewConfig creates and returns an incomplete Config object
func NewConfig() *Config {
	return &Config{
		TLS:                    true,
		Timeout:                60 * time.Second,
		WaitConnectionInterval: 5 * time.Second,
		MaxConnectionAttempts:  0,
	}
}
func (c *Config) Addresses() []string {
	c.adressMux.Lock()
	defer c.adressMux.Unlock()
	return c.addresses
}
func (c *Config) SetAddresses(addresses []string) {
	c.adressMux.Lock()
	defer c.adressMux.Unlock()
	c.addresses = addresses
}

// AsConnectionString constructs a MongoDB connection string from a Config
func (c *Config) AsConnectionString() string {
	var url string
	if c.Scheme != "" {
		url += c.Scheme + "://"
	} else {
		url += "mongodb://"
	}

	if c.Username != nil && *c.Username != "" {
		url += *c.Username
		if c.Password != nil && *c.Password != "" {
			url += ":"
			url += *c.Password
		}
		url += "@"
	}
	url += strings.Join(c.Addresses(), ",")
	url += "/"
	url += c.Database
	if c.TLS {
		url += "?ssl=true"
	} else {
		url += "?ssl=false"
	}
	if c.OptParams != nil && *c.OptParams != "" {
		url += *c.OptParams
	}

	return url
}

// Load a Config with the values provided via a config.Reporter
func (c *Config) Load(configReporter config.Reporter) error {
	if configReporter == nil {
		return errors.New("config reporter is missing")
	}

	addresses := SplitAddresses(configReporter.GetWithDefault("addresses", strings.Join(c.Addresses(), ",")))
	c.SetAddresses(addresses)
	if tlsString, err := configReporter.Get("tls"); err == nil {
		var tls bool
		tls, err = strconv.ParseBool(tlsString)
		if err != nil {
			return errors.New("tls is invalid")
		}
		c.TLS = tls
	}
	c.Scheme = configReporter.GetWithDefault("scheme", c.Scheme)
	c.Database = configReporter.GetWithDefault("database", c.Database)
	c.CollectionPrefix = configReporter.GetWithDefault("collection_prefix", c.CollectionPrefix)
	if username, err := configReporter.Get("username"); err == nil {
		c.Username = pointer.FromString(username)
	}
	if password, err := configReporter.Get("password"); err == nil {
		c.Password = pointer.FromString(password)
	}
	if optParams, err := configReporter.Get("opt_params"); err == nil {
		c.OptParams = pointer.FromString(optParams)
	}
	if timeoutString, err := configReporter.Get("timeout"); err == nil {
		var timeout int64
		timeout, err = strconv.ParseInt(timeoutString, 10, 0)
		if err != nil {
			return errors.New("timeout is invalid")
		}
		c.Timeout = time.Duration(timeout) * time.Second
	}
	if waitConnectionInterval, err := configReporter.Get("wait_connection_interval"); err == nil {
		interval, err := strconv.ParseInt(waitConnectionInterval, 10, 0)
		if err != nil {
			return errors.New("wait_connection_interval is invalid")
		}
		c.WaitConnectionInterval = time.Duration(interval) * time.Second
	}
	if maxConnectionAttempts, err := configReporter.Get("max_connection_attempts"); err == nil {
		max, err := strconv.ParseInt(maxConnectionAttempts, 10, 0)
		if err != nil {
			return errors.New("max_connection_attempts is invalid")
		}
		c.MaxConnectionAttempts = max
	}

	return nil
}

// Validate that all parameters are syntactically valid, that all required parameters are present,
// and the the URL constructed from those parameters is parseable by the Mongo driver
func (c *Config) Validate() error {
	if len(c.Addresses()) == 0 {
		return errors.New("addresses is missing")
	}
	for _, address := range c.Addresses() {
		if address == "" {
			return errors.New("address is missing")
		} else if _, err := url.Parse(address); err != nil {
			return errors.New("address is invalid")
		}
	}
	if c.Database == "" {
		return errors.New("database is missing")
	}
	if c.Timeout <= 0 {
		return errors.New("timeout is invalid")
	}

	if _, err := mgo.ParseURL(c.AsConnectionString()); err != nil {
		return errors.New("URL is unparseable by driver, check validity of optional parameters")
	}

	return nil
}

func SplitAddresses(addresses string) []string {
	return config.SplitTrimCompact(addresses)
}
