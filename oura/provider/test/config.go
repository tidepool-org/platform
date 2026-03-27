package test

import (
	oauthProviderClientTest "github.com/tidepool-org/platform/oauth/provider/client/test"
	ouraProvider "github.com/tidepool-org/platform/oura/provider"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

func RandomPartnerURL() string {
	return testHttp.NewURL().String()
}

func RandomPartnerSecret() string {
	return test.RandomStringFromCharset(test.CharsetAlphaNumeric)
}

func RandomConfig(options ...test.Option) *ouraProvider.Config {
	return &ouraProvider.Config{
		Config:        oauthProviderClientTest.RandomConfig(options...),
		PartnerURL:    RandomPartnerURL(),
		PartnerSecret: RandomPartnerSecret(),
	}
}

func CloneConfig(config *ouraProvider.Config) *ouraProvider.Config {
	if config == nil {
		return nil
	}
	clone := &ouraProvider.Config{
		Config:        oauthProviderClientTest.CloneConfig(config.Config),
		PartnerURL:    config.PartnerURL,
		PartnerSecret: config.PartnerSecret,
	}
	return clone
}

func NewObjectFromConfig(config *ouraProvider.Config, objectFormat test.ObjectFormat) map[string]any {
	if config == nil {
		return nil
	}
	object := map[string]any{}
	if config.Config != nil {
		object = oauthProviderClientTest.NewObjectFromConfig(config.Config, objectFormat)
	}
	object["partner_url"] = test.NewObjectFromString(config.PartnerURL, objectFormat)
	if objectFormat == test.ObjectFormatConfig {
		object["partner_secret"] = test.NewObjectFromString(config.PartnerSecret, objectFormat)
	}
	return object
}
