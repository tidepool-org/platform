package test

import (
	"maps"

	providerSessionWorkTest "github.com/tidepool-org/platform/auth/providersession/work/test"
	dataWork "github.com/tidepool-org/platform/data/work"
	dataWorkTest "github.com/tidepool-org/platform/data/work/test"
	ouraDataWorkHistoric "github.com/tidepool-org/platform/oura/data/work/historic"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	timesTest "github.com/tidepool-org/platform/times/test"
)

func RandomDataTypeNextTokens(options ...test.Option) *dataWork.StringStringMap {
	dataTypeNextTokens := dataWork.StringStringMap{}
	for _, dataType := range test.RandomStringArrayFromArrayWithoutDuplicates(ouraDataWorkHistoric.DataTypes()) {
		if test.IsOptionalPresent(options...) {
			dataTypeNextTokens[dataType] = pointer.From(ouraTest.RandomNextToken())
		} else {
			dataTypeNextTokens[dataType] = nil
		}
	}
	return pointer.From(dataTypeNextTokens)
}

func CloneDataTypeNextTokens(datum *dataWork.StringStringMap) *dataWork.StringStringMap {
	return dataWorkTest.CloneStringStringMap(datum)
}

func NewObjectFromDataTypeNextTokens(datum *dataWork.StringStringMap, format test.ObjectFormat) map[string]any {
	return dataWorkTest.NewObjectFromStringStringMap(datum, format)
}

func RandomMetadata(options ...test.Option) *ouraDataWorkHistoric.Metadata {
	return &ouraDataWorkHistoric.Metadata{
		ProviderSessionMetadata: *providerSessionWorkTest.RandomMetadata(),
		TimeRangeMetadata:       *timesTest.RandomTimeRangeMetadata(options...),
		DataTypeNextTokens:      RandomDataTypeNextTokens(options...),
	}
}

func CloneMetadata(datum *ouraDataWorkHistoric.Metadata) *ouraDataWorkHistoric.Metadata {
	if datum == nil {
		return nil
	}
	return &ouraDataWorkHistoric.Metadata{
		ProviderSessionMetadata: *providerSessionWorkTest.CloneMetadata(&datum.ProviderSessionMetadata),
		TimeRangeMetadata:       *timesTest.CloneTimeRangeMetadata(&datum.TimeRangeMetadata),
		DataTypeNextTokens:      CloneDataTypeNextTokens(datum.DataTypeNextTokens),
	}
}

func NewObjectFromMetadata(datum *ouraDataWorkHistoric.Metadata, format test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	maps.Copy(object, providerSessionWorkTest.NewObjectFromMetadata(&datum.ProviderSessionMetadata, format))
	maps.Copy(object, timesTest.NewObjectFromTimeRangeMetadata(&datum.TimeRangeMetadata, format))
	if datum.DataTypeNextTokens != nil {
		object[ouraDataWorkHistoric.MetadataKeyDataTypeNextTokens] = NewObjectFromDataTypeNextTokens(datum.DataTypeNextTokens, format)
	}
	return object
}
