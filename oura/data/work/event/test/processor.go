package test

import (
	"maps"

	providerSessionWorkTest "github.com/tidepool-org/platform/auth/providersession/work/test"
	ouraDataWorkEvent "github.com/tidepool-org/platform/oura/data/work/event"
	ouraWebhookTest "github.com/tidepool-org/platform/oura/webhook/test"
	"github.com/tidepool-org/platform/test"
)

func RandomMetadata(options ...test.Option) *ouraDataWorkEvent.Metadata {
	return &ouraDataWorkEvent.Metadata{
		ProviderSessionMetadata: *providerSessionWorkTest.RandomMetadata(options...),
		EventMetadata:           *ouraWebhookTest.RandomEventMetadata(options...),
	}
}

func CloneMetadata(datum *ouraDataWorkEvent.Metadata) *ouraDataWorkEvent.Metadata {
	if datum == nil {
		return nil
	}
	return &ouraDataWorkEvent.Metadata{
		ProviderSessionMetadata: *providerSessionWorkTest.CloneMetadata(&datum.ProviderSessionMetadata),
		EventMetadata:           *ouraWebhookTest.CloneEventMetadata(&datum.EventMetadata),
	}
}

func NewObjectFromMetadata(datum *ouraDataWorkEvent.Metadata, format test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	maps.Copy(object, providerSessionWorkTest.NewObjectFromMetadata(&datum.ProviderSessionMetadata, format))
	maps.Copy(object, ouraWebhookTest.NewObjectFromEventMetadata(&datum.EventMetadata, format))
	return object
}
