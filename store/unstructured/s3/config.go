package s3

import (
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
)

type Config struct {
	Bucket string
	Prefix string
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Load(configReporter config.Reporter) error {
	if configReporter == nil {
		return errors.New("config reporter is missing")
	}

	c.Bucket = configReporter.GetWithDefault("bucket", c.Bucket)
	c.Prefix = configReporter.GetWithDefault("prefix", c.Prefix)
	return nil
}

func (c *Config) Validate() error {
	if c.Bucket == "" {
		return errors.New("bucket is missing")
	}
	if !storeUnstructured.IsValidKey(c.Prefix) {
		return errors.New("prefix is invalid")
	}

	return nil
}
