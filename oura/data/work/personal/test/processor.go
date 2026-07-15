package test

import (
	"maps"

	providerSessionWorkTest "github.com/tidepool-org/platform/auth/providersession/work/test"
	cryptoTest "github.com/tidepool-org/platform/crypto/test"
	ouraDataWorkPersonal "github.com/tidepool-org/platform/oura/data/work/personal"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomMetadata(options ...test.Option) *ouraDataWorkPersonal.Metadata {
	return &ouraDataWorkPersonal.Metadata{
		ProviderSessionMetadata: *providerSessionWorkTest.RandomMetadata(),
		PreviousHash:            test.RandomOptional(cryptoTest.RandomBase64EncodedSHA256Hash, options...),
	}
}

func CloneMetadata(datum *ouraDataWorkPersonal.Metadata) *ouraDataWorkPersonal.Metadata {
	if datum == nil {
		return nil
	}
	return &ouraDataWorkPersonal.Metadata{
		ProviderSessionMetadata: *providerSessionWorkTest.CloneMetadata(&datum.ProviderSessionMetadata),
		PreviousHash:            pointer.Clone(datum.PreviousHash),
	}
}

func NewObjectFromMetadata(datum *ouraDataWorkPersonal.Metadata, format test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	maps.Copy(object, providerSessionWorkTest.NewObjectFromMetadata(&datum.ProviderSessionMetadata, format))
	if datum.PreviousHash != nil {
		object[ouraDataWorkPersonal.MetadataKeyPreviousHash] = test.NewObjectFromString(*datum.PreviousHash, format)
	}
	return object
}
