package test

import (
	"maps"

	ouraData "github.com/tidepool-org/platform/oura/data"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	"github.com/tidepool-org/platform/test"
	timesTest "github.com/tidepool-org/platform/times/test"
)

func RandomMetadata(options ...test.Option) *ouraData.Metadata {
	datum := &ouraData.Metadata{
		DataType: ouraTest.RandomDataType(),
	}
	if test.RandomBool() {
		datum.EventMetadata = *ouraTest.RandomEventMetadata(options...)
	} else {
		datum.TimeRangeMetadata = *timesTest.RandomTimeRangeMetadata(options...)
	}
	return datum
}

func CloneMetadata(datum *ouraData.Metadata) *ouraData.Metadata {
	if datum == nil {
		return nil
	}
	return &ouraData.Metadata{
		DataType:          datum.DataType,
		EventMetadata:     *ouraTest.CloneEventMetadata(&datum.EventMetadata),
		TimeRangeMetadata: *timesTest.CloneTimeRangeMetadata(&datum.TimeRangeMetadata),
	}
}

func NewObjectFromMetadata(datum *ouraData.Metadata, format test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.DataType != "" {
		object[ouraData.MetadataKeyDataType] = test.NewObjectFromString(datum.DataType, format)
	}
	maps.Copy(object, ouraTest.NewObjectFromEventMetadata(&datum.EventMetadata, format))
	maps.Copy(object, timesTest.NewObjectFromTimeRangeMetadata(&datum.TimeRangeMetadata, format))
	return object
}
