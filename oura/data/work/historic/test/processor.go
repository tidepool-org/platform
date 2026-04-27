package test

import (
	"maps"

	providerSessionWorkTest "github.com/tidepool-org/platform/auth/providersession/work/test"
	ouraDataWorkHistoric "github.com/tidepool-org/platform/oura/data/work/historic"
	"github.com/tidepool-org/platform/test"
	timesTest "github.com/tidepool-org/platform/times/test"
)

func RandomMetadata(options ...test.Option) *ouraDataWorkHistoric.Metadata {
	return &ouraDataWorkHistoric.Metadata{
		ProviderSessionMetadata: *providerSessionWorkTest.RandomMetadata(options...),
		TimeRangeMetadata:       *timesTest.RandomTimeRangeMetadata(options...),
	}
}

func CloneMetadata(datum *ouraDataWorkHistoric.Metadata) *ouraDataWorkHistoric.Metadata {
	if datum == nil {
		return nil
	}
	return &ouraDataWorkHistoric.Metadata{
		ProviderSessionMetadata: *providerSessionWorkTest.CloneMetadata(&datum.ProviderSessionMetadata),
		TimeRangeMetadata:       *timesTest.CloneTimeRangeMetadata(&datum.TimeRangeMetadata),
	}
}

func NewObjectFromMetadata(datum *ouraDataWorkHistoric.Metadata, format test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	maps.Copy(object, providerSessionWorkTest.NewObjectFromMetadata(&datum.ProviderSessionMetadata, format))
	maps.Copy(object, timesTest.NewObjectFromTimeRangeMetadata(&datum.TimeRangeMetadata, format))
	return object
}
