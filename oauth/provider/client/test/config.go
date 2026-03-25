package test

import (
	clientTest "github.com/tidepool-org/platform/client/test"
	oauthProviderClient "github.com/tidepool-org/platform/oauth/provider/client"
	oauthProviderTest "github.com/tidepool-org/platform/oauth/provider/test"
	"github.com/tidepool-org/platform/test"
)

func RandomConfig(options ...test.Option) *oauthProviderClient.Config {
	return &oauthProviderClient.Config{
		Provider: oauthProviderTest.RandomConfig(options...),
		Client:   clientTest.RandomConfig(options...),
	}
}

func CloneConfig(config *oauthProviderClient.Config) *oauthProviderClient.Config {
	if config == nil {
		return nil
	}
	clone := &oauthProviderClient.Config{
		Provider: oauthProviderTest.CloneConfig(config.Provider),
		Client:   clientTest.CloneConfig(config.Client),
	}
	return clone
}

func NewObjectFromConfig(config *oauthProviderClient.Config, objectFormat test.ObjectFormat) map[string]any {
	if config == nil {
		return nil
	}
	object := map[string]any{}
	if config.Provider != nil {
		object = oauthProviderTest.NewObjectFromConfig(config.Provider, objectFormat)
	}
	if config.Client != nil {
		object["client"] = clientTest.NewObjectFromConfig(config.Client, objectFormat)
	}
	return object
}
