package provider

import (
	"github.com/tidepool-org/platform/config"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
)

type ProviderSessionClient any

type DataSourceClient any

type WorkClient any

type ProviderDependencies struct {
	ConfigReporter        config.Reporter
	ProviderSessionClient ProviderSessionClient
	DataSourceClient      DataSourceClient
	WorkClient            WorkClient
}

type Provider struct {
	*oauthProvider.Provider
}

func NewProvider(providerDependencies ProviderDependencies) (*Provider, error) {
	return nil, nil
}
