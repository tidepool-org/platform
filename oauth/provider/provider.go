package provider

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oauth"
)

const ProviderType = "oauth"

type Provider struct {
	name      string
	config    *oauth2.Config
	stateSalt string
}

func NewProvider(name string, configReporter config.Reporter) (*Provider, error) {
	if name == "" {
		return nil, errors.New("name is missing")
	}
	if configReporter == nil {
		return nil, errors.New("config reporter is missing")
	}

	cfg := &oauth2.Config{}
	cfg.ClientID = configReporter.GetWithDefault("client_id", "")
	if cfg.ClientID == "" {
		return nil, errors.New("client id is missing")
	}
	cfg.ClientSecret = configReporter.GetWithDefault("client_secret", "")
	if cfg.ClientSecret == "" {
		return nil, errors.New("client secret is missing")
	}
	cfg.Endpoint.AuthURL = configReporter.GetWithDefault("authorize_url", "")
	if cfg.Endpoint.AuthURL == "" {
		return nil, errors.New("authorize url is missing")
	}
	cfg.Endpoint.TokenURL = configReporter.GetWithDefault("token_url", "")
	if cfg.Endpoint.TokenURL == "" {
		return nil, errors.New("token url is missing")
	}
	cfg.RedirectURL = configReporter.GetWithDefault("redirect_url", "")
	if cfg.RedirectURL == "" {
		return nil, errors.New("redirect url is missing")
	}
	cfg.Scopes = SplitScopes(configReporter.GetWithDefault("scopes", ""))

	authStyleInParams := configReporter.GetWithDefault("auth_style_in_params", "")
	if authStyleInParams == "true" {
		cfg.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	}

	stateSalt := configReporter.GetWithDefault("state_salt", "")
	if stateSalt == "" {
		return nil, errors.New("state salt is missing")
	}

	return &Provider{
		name:      name,
		config:    cfg,
		stateSalt: stateSalt,
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

func (p *Provider) TokenSource(ctx context.Context, token *oauth.Token) (oauth2.TokenSource, error) {
	if token == nil {
		return nil, errors.New("token is missing")
	}

	tknSrc := p.config.TokenSource(ctx, token.RawToken())
	if tknSrc == nil {
		return nil, errors.New("unable to create token source")
	}

	return tknSrc, nil
}

func (p *Provider) CalculateStateForRestrictedToken(restrictedToken string) string {
	return crypto.HexEncodedMD5Hash(fmt.Sprintf("%s:%s:%s:%s", p.Type(), p.Name(), restrictedToken, p.stateSalt))
}

func (p *Provider) GetAuthorizationCodeURLWithState(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *Provider) ExchangeAuthorizationCodeForToken(ctx context.Context, authorizationCode string) (*oauth.Token, error) {
	token, err := p.config.Exchange(ctx, authorizationCode)
	if err != nil {
		return nil, errors.Wrap(err, "unable to exchange authorization code for token")
	}

	return oauth.NewTokenFromRawToken(token)
}

func SplitScopes(scopes string) []string {
	return config.SplitTrimCompact(scopes)
}
