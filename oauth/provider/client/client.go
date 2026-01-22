package client

import (
	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/tidepool-org/platform/errors"
	oauthClient "github.com/tidepool-org/platform/oauth/client"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
)

type Provider struct {
	*oauthProvider.Provider
	*oauthClient.Client
}

func New(name string, config *Config, jwks jwk.Set) (*Provider, error) {
	if name == "" {
		return nil, errors.New("name is missing")
	}
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	provider, err := oauthProvider.New(name, config.Provider, jwks)
	if err != nil {
		return nil, err
	}
	client, err := oauthClient.New(config.Client, provider)
	if err != nil {
		return nil, err
	}

	return &Provider{
		Provider: provider,
		Client:   client,
	}, nil
}
