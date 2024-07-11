package mongo

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"

	platformConfig "github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
)

func NewConfig() *Config {
	return &Config{
		TLS:     true,
		Timeout: 30 * time.Second,
	}
}

func LoadConfig() (*Config, error) {
	cfg := NewConfig()
	err := cfg.Load()
	return cfg, err
}

// Config describe parameters need to make a connection to a Mongo database
type Config struct {
	Scheme           string        `json:"scheme" envconfig:"TIDEPOOL_STORE_SCHEME"`
	Addresses        []string      `json:"addresses" envconfig:"TIDEPOOL_STORE_ADDRESSES" required:"true"`
	TLS              bool          `json:"tls" envconfig:"TIDEPOOL_STORE_TLS" default:"true"`
	Database         string        `json:"database" envconfig:"TIDEPOOL_STORE_DATABASE" required:"true"`
	CollectionPrefix string        `json:"collectionPrefix" envconfig:"TIDEPOOL_STORE_COLLECTION_PREFIX"`
	Username         *string       `json:"-" envconfig:"TIDEPOOL_STORE_USERNAME"`
	Password         *string       `json:"-" envconfig:"TIDEPOOL_STORE_PASSWORD"`
	Timeout          time.Duration `json:"timeout" envconfig:"TIDEPOOL_STORE_TIMEOUT" default:"60s"`
	OptParams        *string       `json:"optParams" envconfig:"TIDEPOOL_STORE_OPT_PARAMS"`
	AppName          *string       `json:"appName" envconfig:"TIDEPOOL_STORE_APP_NAME"`
}

// AsConnectionString constructs a MongoDB connection string from a Config
func (c *Config) AsConnectionString() string {
	var connectionString string
	if c.Scheme != "" {
		connectionString += c.Scheme + "://"
	} else {
		connectionString += "mongodb://"
	}

	if c.Username != nil && *c.Username != "" {
		connectionString += *c.Username
		if c.Password != nil && *c.Password != "" {
			connectionString += ":"
			connectionString += *c.Password
		}
		connectionString += "@"
	}
	connectionString += strings.Join(c.Addresses, ",")
	connectionString += "/"
	connectionString += c.Database
	if c.TLS {
		connectionString += "?ssl=true"
	} else {
		connectionString += "?ssl=false"
	}
	if c.OptParams != nil && *c.OptParams != "" {
		connectionString += fmt.Sprintf("&%s", *c.OptParams)
	}
	if c.AppName != nil && *c.AppName != "" {
		connectionString += fmt.Sprintf("&appName=%s", url.QueryEscape(*c.AppName))
	}

	return connectionString
}

func (c *Config) Load() error {
	return envconfig.Process("", c)
}

func (c *Config) SetDatabaseFromReporter(configReporter platformConfig.Reporter) error {
	var err error
	c.Database, err = configReporter.Get("database")
	if err != nil {
		return err
	}
	return nil
}

// Validate that all parameters are syntactically valid, that all required parameters are present,
// and the the URL constructed from those parameters is parseable by the Mongo driver
func (c *Config) Validate() error {
	if len(c.Addresses) == 0 {
		return errors.New("addresses is missing")
	}
	for _, address := range c.Addresses {
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

	if _, err := connstring.Parse(c.AsConnectionString()); err != nil {
		return errors.Wrap(err, "URL is unparseable by driver, check validity of optional parameters")
	}

	return nil
}
