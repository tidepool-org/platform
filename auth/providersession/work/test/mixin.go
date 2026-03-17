package test

import (
	providerSessionWork "github.com/tidepool-org/platform/auth/providersession/work"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomMetadata(options ...test.Option) *providerSessionWork.Metadata {
	return &providerSessionWork.Metadata{
		ProviderSessionID: test.RandomOptional(authTest.RandomProviderSessionID, options...),
	}
}

func CloneMetadata(datum *providerSessionWork.Metadata) *providerSessionWork.Metadata {
	if datum == nil {
		return nil
	}
	return &providerSessionWork.Metadata{
		ProviderSessionID: pointer.CloneString(datum.ProviderSessionID),
	}
}

func NewObjectFromMetadata(datum *providerSessionWork.Metadata, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.ProviderSessionID != nil {
		object[providerSessionWork.MetadataKeyProviderSessionID] = test.NewObjectFromString(*datum.ProviderSessionID, objectFormat)
	}
	return object
}
