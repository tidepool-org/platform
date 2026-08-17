package client

import (
	"net/http"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/errors"
	oauthClient "github.com/tidepool-org/platform/oauth/client"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
)

type Provider struct {
	*oauthProvider.Provider
	*oauthClient.Client
}

func New(name string, config *Config, httpClient *http.Client, jwks jwk.Set) (*Provider, error) {
	return NewWithErrorParser(name, config, httpClient, jwks, nil)
}

func NewWithErrorParser(name string, config *Config, httpClient *http.Client, jwks jwk.Set, errorResponseParser client.ErrorResponseParser) (*Provider, error) {
	if name == "" {
		return nil, errors.New("name is missing")
	}
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	prvdr, err := oauthProvider.New(name, config.ProviderConfig, jwks)
	if err != nil {
		return nil, err
	}
	clnt, err := oauthClient.NewWithErrorParser(config.ClientConfig, httpClient, prvdr, errorResponseParser)
	if err != nil {
		return nil, err
	}

	return &Provider{
		Provider: prvdr,
		Client:   clnt,
	}, nil
}
