package test

import (
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

func RandomConfig(options ...test.Option) *client.Config {
	return &client.Config{
		Address:   testHttp.NewURL().String(),
		UserAgent: test.RandomStringFromCharset(test.CharsetAlphaNumeric),
	}
}

func CloneConfig(config *client.Config) *client.Config {
	if config == nil {
		return nil
	}
	clone := &client.Config{
		Address:   config.Address,
		UserAgent: config.UserAgent,
	}
	return clone
}

func NewObjectFromConfig(config *client.Config, objectFormat test.ObjectFormat) map[string]any {
	if config == nil {
		return nil
	}
	object := map[string]any{}
	object["address"] = test.NewObjectFromString(config.Address, objectFormat)
	object["user_agent"] = test.NewObjectFromString(config.UserAgent, objectFormat)
	return object

}
