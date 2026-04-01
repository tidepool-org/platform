package provider

import (
	"net/url"
	"strconv"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
)

type Config struct {
	ClientID          string   `json:"client_id,omitempty"`
	ClientSecret      string   `json:"client_secret,omitempty"`
	AcceptURL         *string  `json:"accept_url,omitempty"`
	AuthorizeURL      string   `json:"authorize_url,omitempty"`
	RedirectURL       string   `json:"redirect_url,omitempty"`
	TokenURL          string   `json:"token_url,omitempty"`
	RevokeURL         *string  `json:"revoke_url,omitempty"`
	Scopes            []string `json:"scopes,omitempty"`
	AuthStyleInParams bool     `json:"auth_style_in_params,omitempty"`
	CookieDisabled    bool     `json:"cookie_disabled,omitempty"`
	StateSalt         *string  `json:"state_salt,omitempty"`
}

func NewConfigWithConfigReporter(configReporter config.Reporter) (*Config, error) {
	config := NewConfig()
	if err := config.LoadFromConfigReporter(configReporter); err != nil {
		return nil, err
	}
	return config, nil
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) LoadFromConfigReporter(configReporter config.Reporter) error {
	if configReporter == nil {
		return errors.New("config reporter is missing")
	}
	c.ClientID = configReporter.GetWithDefault("client_id", c.ClientID)
	c.ClientSecret = configReporter.GetWithDefault("client_secret", c.ClientSecret)
	if acceptURL, err := configReporter.Get("accept_url"); err == nil && acceptURL != "" {
		c.AcceptURL = &acceptURL
	}
	c.AuthorizeURL = configReporter.GetWithDefault("authorize_url", c.AuthorizeURL)
	c.RedirectURL = configReporter.GetWithDefault("redirect_url", c.RedirectURL)
	c.TokenURL = configReporter.GetWithDefault("token_url", c.ClientID)
	if revokeURL, err := configReporter.Get("revoke_url"); err == nil && revokeURL != "" {
		c.RevokeURL = &revokeURL
	}
	if scopesString, err := configReporter.Get("scopes"); err == nil {
		if scopes, err := auth.ParseScope(scopesString); err != nil {
			return errors.Wrap(err, "scopes is invalid")
		} else {
			c.Scopes = scopes
		}
	}
	if authStyleInParams, err := strconv.ParseBool(configReporter.GetWithDefault("auth_style_in_params", strconv.FormatBool(c.AuthStyleInParams))); err != nil {
		return errors.New("auth style in params is invalid")
	} else {
		c.AuthStyleInParams = authStyleInParams
	}
	if cookieDisabled, err := strconv.ParseBool(configReporter.GetWithDefault("cookie_disabled", strconv.FormatBool(c.CookieDisabled))); err != nil {
		return errors.New("cookie disabled is invalid")
	} else {
		c.CookieDisabled = cookieDisabled
	}
	if stateSalt, err := configReporter.Get("state_salt"); err == nil && stateSalt != "" {
		c.StateSalt = pointer.FromString(stateSalt)
	}
	return nil
}

func (c *Config) Validate() error {
	if c.ClientID == "" {
		return errors.New("client id is empty")
	}
	if c.ClientSecret == "" {
		return errors.New("client secret is empty")
	}
	if c.AcceptURL != nil {
		if _, err := url.Parse(*c.AcceptURL); err != nil {
			return errors.New("accept url is invalid")
		}
	}
	if c.AuthorizeURL == "" {
		return errors.New("authorize url is empty")
	}
	if c.RedirectURL == "" {
		return errors.New("redirect url is empty")
	}
	if c.TokenURL == "" {
		return errors.New("token url is empty")
	}
	if c.RevokeURL != nil {
		if _, err := url.Parse(*c.RevokeURL); err != nil {
			return errors.New("revoke url is invalid")
		}
	}
	if !c.CookieDisabled {
		if c.StateSalt == nil {
			return errors.New("state salt is missing")
		} else if *c.StateSalt == "" {
			return errors.New("state salt is empty")
		}
	}
	return nil
}
