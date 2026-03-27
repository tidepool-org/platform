package test

import (
	clientTest "github.com/tidepool-org/platform/client/test"
	oauthProviderClient "github.com/tidepool-org/platform/oauth/provider/client"
	oauthProviderTest "github.com/tidepool-org/platform/oauth/provider/test"
	"github.com/tidepool-org/platform/test"
)

func RandomConfig(options ...test.Option) *oauthProviderClient.Config {
	return &oauthProviderClient.Config{
		ProviderConfig: oauthProviderTest.RandomConfig(options...),
		ClientConfig:   clientTest.RandomConfig(options...),
	}
}

func CloneConfig(config *oauthProviderClient.Config) *oauthProviderClient.Config {
	if config == nil {
		return nil
	}
	clone := &oauthProviderClient.Config{
		ProviderConfig: oauthProviderTest.CloneConfig(config.ProviderConfig),
		ClientConfig:   clientTest.CloneConfig(config.ClientConfig),
	}
	return clone
}

func NewObjectFromConfig(config *oauthProviderClient.Config, objectFormat test.ObjectFormat) map[string]any {
	if config == nil {
		return nil
	}
	object := map[string]any{}
	if config.ProviderConfig != nil {
		object = oauthProviderTest.NewObjectFromConfig(config.ProviderConfig, objectFormat)
	}
	if config.ClientConfig != nil {
		object["client"] = clientTest.NewObjectFromConfig(config.ClientConfig, objectFormat)
	}
	return object
}
