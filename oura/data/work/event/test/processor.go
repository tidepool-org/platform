package test

import (
	"maps"

	providerSessionWorkTest "github.com/tidepool-org/platform/auth/providersession/work/test"
	ouraDataWorkEvent "github.com/tidepool-org/platform/oura/data/work/event"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	"github.com/tidepool-org/platform/test"
)

func RandomMetadata(options ...test.Option) *ouraDataWorkEvent.Metadata {
	return &ouraDataWorkEvent.Metadata{
		ProviderSessionMetadata: *providerSessionWorkTest.RandomMetadata(),
		EventMetadata:           *ouraTest.RandomEventMetadata(options...),
	}
}

func CloneMetadata(datum *ouraDataWorkEvent.Metadata) *ouraDataWorkEvent.Metadata {
	if datum == nil {
		return nil
	}
	return &ouraDataWorkEvent.Metadata{
		ProviderSessionMetadata: *providerSessionWorkTest.CloneMetadata(&datum.ProviderSessionMetadata),
		EventMetadata:           *ouraTest.CloneEventMetadata(&datum.EventMetadata),
	}
}

func NewObjectFromMetadata(datum *ouraDataWorkEvent.Metadata, format test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	maps.Copy(object, providerSessionWorkTest.NewObjectFromMetadata(&datum.ProviderSessionMetadata, format))
	maps.Copy(object, ouraTest.NewObjectFromEventMetadata(&datum.EventMetadata, format))
	return object
}
