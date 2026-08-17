package test

import (
	"maps"

	providerSessionWorkTest "github.com/tidepool-org/platform/auth/providersession/work/test"
	ouraUserWorkSetup "github.com/tidepool-org/platform/oura/user/work/setup"
	"github.com/tidepool-org/platform/test"
)

func RandomMetadata(options ...test.Option) *ouraUserWorkSetup.Metadata {
	return &ouraUserWorkSetup.Metadata{
		ProviderSessionMetadata: *providerSessionWorkTest.RandomMetadata(),
	}
}

func CloneMetadata(datum *ouraUserWorkSetup.Metadata) *ouraUserWorkSetup.Metadata {
	if datum == nil {
		return nil
	}
	return &ouraUserWorkSetup.Metadata{
		ProviderSessionMetadata: *providerSessionWorkTest.CloneMetadata(&datum.ProviderSessionMetadata),
	}
}

func NewObjectFromMetadata(datum *ouraUserWorkSetup.Metadata, format test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	maps.Copy(object, providerSessionWorkTest.NewObjectFromMetadata(&datum.ProviderSessionMetadata, format))
	return object
}
