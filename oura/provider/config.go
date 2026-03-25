package provider

import (
	"net/url"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	oauthProviderClient "github.com/tidepool-org/platform/oauth/provider/client"
)

type Config struct {
	*oauthProviderClient.Config
	PartnerURL    string `json:"partner_url,omitempty"`
	PartnerSecret string `json:"-"`
}

func NewConfigWithConfigReporter(configReporter config.Reporter) (*Config, error) {
	config := NewConfig()
	if err := config.LoadFromConfigReporter(configReporter); err != nil {
		return nil, err
	}
	return config, nil
}

func NewConfig() *Config {
	return &Config{
		Config: oauthProviderClient.NewConfig(),
	}
}

func (c *Config) LoadFromConfigReporter(configReporter config.Reporter) error {
	if c.Config != nil {
		if err := c.Config.LoadFromConfigReporter(configReporter); err != nil {
			return err
		}
	}
	c.PartnerSecret = configReporter.GetWithDefault("partner_secret", c.PartnerSecret)
	c.PartnerURL = configReporter.GetWithDefault("partner_url", c.PartnerURL)
	return nil
}

func (c *Config) Validate() error {
	if c.Config == nil {
		return errors.New("config is missing")
	} else if err := c.Config.Validate(); err != nil {
		return errors.Wrap(err, "config is invalid")
	} else if c.Provider.AcceptURL == nil {
		return errors.Wrap(errors.Wrap(errors.New("accept url is missing"), "provider is invalid"), "config is invalid")
	} else if c.Provider.RevokeURL == nil {
		return errors.Wrap(errors.Wrap(errors.New("revoke url is missing"), "provider is invalid"), "config is invalid")
	}
	if c.PartnerURL == "" {
		return errors.New("partner url is missing")
	} else if _, err := url.Parse(c.PartnerURL); err != nil {
		return errors.New("partner url is invalid")
	}
	if c.PartnerSecret == "" {
		return errors.New("partner secret is missing")
	}
	return nil
}
