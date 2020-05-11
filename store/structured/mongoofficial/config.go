package mongoofficial

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"

	"github.com/tidepool-org/platform/errors"

	"go.uber.org/fx"
)

func NewConfig(lifecycle fx.Lifecycle) *Config {
	cfg := &Config{}

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := cfg.Load(); err != nil {
				return err
			}

			return cfg.Validate()
		},
	})

	return cfg
}

//Config describe parameters need to make a connection to a Mongo database
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
		connectionString += *c.OptParams
	}

	return connectionString
}

func (c *Config) Load() error {
	return envconfig.Process("", c)
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
			fmt.Fprintf(os.Stdout, "%v", address)
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
		return errors.New("URL is unparseable by driver, check validity of optional parameters")
	}

	return nil
}
