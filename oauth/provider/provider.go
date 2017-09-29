package provider

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/errors"
)

const ProviderType = "oauth"

type Provider struct {
	name         string
	clientID     string
	clientSecret string
	authorizeURL string
	tokenURL     string
	redirectURL  string
	scopes       []string
	stateSalt    string
}

func New(name string, configReporter config.Reporter) (*Provider, error) {
	if name == "" {
		return nil, errors.New("name is missing")
	}
	if configReporter == nil {
		return nil, errors.New("config reporter is missing")
	}

	clientID := configReporter.GetWithDefault("client_id", "")
	if clientID == "" {
		return nil, errors.New("client id is missing")
	}
	clientSecret := configReporter.GetWithDefault("client_secret", "")
	if clientSecret == "" {
		return nil, errors.New("client secret is missing")
	}
	authorizeURL := configReporter.GetWithDefault("authorize_url", "")
	if authorizeURL == "" {
		return nil, errors.New("authorize url is missing")
	}
	tokenURL := configReporter.GetWithDefault("token_url", "")
	if tokenURL == "" {
		return nil, errors.New("token url is missing")
	}
	redirectURL := configReporter.GetWithDefault("redirect_url", "")
	if redirectURL == "" {
		return nil, errors.New("redirect url is missing")
	}
	scopes := SplitScopes(configReporter.GetWithDefault("scopes", ""))
	if len(scopes) == 0 {
		return nil, errors.New("scopes is missing")
	}
	stateSalt := configReporter.GetWithDefault("state_salt", "")
	if stateSalt == "" {
		return nil, errors.New("state salt is missing")
	}

	return &Provider{
		name:         name,
		clientID:     clientID,
		clientSecret: clientSecret,
		authorizeURL: authorizeURL,
		tokenURL:     tokenURL,
		redirectURL:  redirectURL,
		scopes:       scopes,
		stateSalt:    stateSalt,
	}, nil
}

func (p *Provider) Type() string {
	return ProviderType
}

func (p *Provider) Name() string {
	return p.name
}

func (p *Provider) OnCreate(ctx context.Context, userID string, providerSessionID string) error {
	return nil
}

func (p *Provider) OnDelete(ctx context.Context, userID string, providerSessionID string) error {
	return nil
}

func (p *Provider) Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.clientID,
		ClientSecret: p.clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  p.authorizeURL,
			TokenURL: p.tokenURL,
		},
		RedirectURL: p.redirectURL,
		Scopes:      p.scopes,
	}
}

func (p *Provider) State(restrictedToken string) string {
	return crypto.HashWithMD5(fmt.Sprintf("%s:%s:%s:%s", p.Type(), p.Name(), restrictedToken, p.stateSalt))
}

func SplitScopes(scopes string) []string {
	return config.SplitTrimCompact(scopes)
}
