package provider

import (
	"context"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
)

const ClientCredentialsProviderType = "oauth_client_credentials"

type ClientCredentialsProvider struct {
	name   string
	config *clientcredentials.Config
}

func NewClientCredentialsProvider(name string, configReporter config.Reporter) (*ClientCredentialsProvider, error) {
	if name == "" {
		return nil, errors.New("name is missing")
	}
	if configReporter == nil {
		return nil, errors.New("config reporter is missing")
	}

	cfg := &clientcredentials.Config{}
	cfg.ClientID = configReporter.GetWithDefault("client_id", "")
	if cfg.ClientID == "" {
		return nil, errors.New("client id is missing")
	}
	cfg.ClientSecret = configReporter.GetWithDefault("client_secret", "")
	if cfg.ClientSecret == "" {
		return nil, errors.New("client secret is missing")
	}
	cfg.TokenURL = configReporter.GetWithDefault("token_url", "")
	if cfg.TokenURL == "" {
		return nil, errors.New("token url is missing")
	}
	cfg.Scopes = SplitScopes(configReporter.GetWithDefault("scopes", ""))
	if str := configReporter.GetWithDefault("endpoint_params", ""); str != "" {
		endpointParams, err := url.ParseQuery(str)
		if err != nil {
			return nil, errors.New("endpoint params is invalid")
		}
		cfg.EndpointParams = endpointParams
	}

	return &ClientCredentialsProvider{
		name:   name,
		config: cfg,
	}, nil
}

func (c *ClientCredentialsProvider) Type() string {
	return ClientCredentialsProviderType
}

func (c *ClientCredentialsProvider) Name() string {
	return c.name
}

func (c *ClientCredentialsProvider) OnCreate(ctx context.Context, providerSession *auth.ProviderSession) error {
	return nil
}

func (c *ClientCredentialsProvider) OnDelete(ctx context.Context, providerSession *auth.ProviderSession) error {
	return nil
}

func (c *ClientCredentialsProvider) TokenSource(ctx context.Context, token *auth.OAuthToken) (oauth2.TokenSource, error) {
	if token != nil {
		return nil, errors.New("token is not missing")
	}

	tknSrc := c.config.TokenSource(ctx)
	if tknSrc == nil {
		return nil, errors.New("unable to create token source")
	}

	return tknSrc, nil
}
