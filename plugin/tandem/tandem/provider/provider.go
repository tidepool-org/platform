package provider

import (
	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/tidepool-org/platform/config"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"

	tandemClient "github.com/tidepool-org/platform-plugin-tandem/tandem/client"
)

type ProviderSessionClient any

type DataSourceClient any

type WorkClient any

type ProviderDependencies struct {
	ConfigReporter        config.Reporter
	ClientConfig          *tandemClient.Config
	ProviderSessionClient ProviderSessionClient
	DataSourceClient      DataSourceClient
	WorkClient            WorkClient
	JWKS                  jwk.Set
}

type Provider struct {
	*oauthProvider.Provider
}

func New(providerDependencies ProviderDependencies) (*Provider, error) {
	return nil, nil
}
