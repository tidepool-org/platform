package test

import (
	"maps"

	authTest "github.com/tidepool-org/platform/auth/test"
	ouraDataWork "github.com/tidepool-org/platform/oura/data/work"
	ouraWebhookTest "github.com/tidepool-org/platform/oura/webhook/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	timesTest "github.com/tidepool-org/platform/times/test"
)

func RandomMetadata(options ...test.Option) *ouraDataWork.Metadata {
	datum := &ouraDataWork.Metadata{
		Scope: test.RandomOptional(authTest.RandomScope, options...),
	}
	if test.RandomBool() {
		datum.TimeRangeMetadata = *timesTest.RandomTimeRangeMetadata()
	} else {
		datum.EventMetadata = *ouraWebhookTest.RandomEventMetadata()
	}
	return datum
}

func CloneMetadata(datum *ouraDataWork.Metadata) *ouraDataWork.Metadata {
	if datum == nil {
		return nil
	}
	return &ouraDataWork.Metadata{
		Scope:             pointer.Clone(datum.Scope),
		TimeRangeMetadata: *timesTest.CloneTimeRangeMetadata(&datum.TimeRangeMetadata),
		EventMetadata:     *ouraWebhookTest.CloneEventMetadata(&datum.EventMetadata),
	}
}

func NewObjectFromMetadata(datum *ouraDataWork.Metadata, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.Scope != nil {
		object[ouraDataWork.MetadataKeyScope] = authTest.NewObjectFromScope(*datum.Scope, objectFormat)
	}
	maps.Copy(object, timesTest.NewObjectFromTimeRangeMetadata(&datum.TimeRangeMetadata, objectFormat))
	maps.Copy(object, ouraWebhookTest.NewObjectFromEventMetadata(&datum.EventMetadata, objectFormat))
	return object
}
