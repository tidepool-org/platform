package provider

import (
	"github.com/tidepool-org/platform/config"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
)

type DataSourceClient interface{}

type WorkClient interface{}

type ProviderDependencies struct {
	ConfigReporter   config.Reporter
	DataSourceClient DataSourceClient
	WorkClient       WorkClient
}

type Provider struct {
	*oauthProvider.Provider
}

func NewProvider(providerDependencies ProviderDependencies) (*Provider, error) {
	return nil, nil
}
