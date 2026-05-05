package test

import (
	ouraWebhookWorkSubscribe "github.com/tidepool-org/platform/oura/webhook/work/subscribe"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomOverride() string {
	return test.RandomStringFromArray(ouraWebhookWorkSubscribe.Overrides())
}

func RandomMetadata(options ...test.Option) *ouraWebhookWorkSubscribe.Metadata {
	return &ouraWebhookWorkSubscribe.Metadata{
		Override: test.RandomOptional(RandomOverride, options...),
	}
}

func CloneMetadata(datum *ouraWebhookWorkSubscribe.Metadata) *ouraWebhookWorkSubscribe.Metadata {
	if datum == nil {
		return nil
	}
	return &ouraWebhookWorkSubscribe.Metadata{
		Override: pointer.Clone(datum.Override),
	}
}

func NewObjectFromMetadata(datum *ouraWebhookWorkSubscribe.Metadata, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.Override != nil {
		object[ouraWebhookWorkSubscribe.MetadataKeyOverride] = test.NewObjectFromString(*datum.Override, objectFormat)
	}
	return object
}
