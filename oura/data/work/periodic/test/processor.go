package test

import (
	"maps"

	providerSessionWorkTest "github.com/tidepool-org/platform/auth/providersession/work/test"
	dataWork "github.com/tidepool-org/platform/data/work"
	dataWorkTest "github.com/tidepool-org/platform/data/work/test"
	ouraDataWorkPeriodic "github.com/tidepool-org/platform/oura/data/work/periodic"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomDataTypeNextTokens(options ...test.Option) *dataWork.StringStringMap {
	dataTypeNextTokens := dataWork.StringStringMap{}
	for _, dataType := range test.RandomStringArrayFromArrayWithoutDuplicates(ouraDataWorkPeriodic.DataTypes()) {
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

func RandomDataTypeStartTimes(options ...test.Option) *dataWork.StringTimeMap {
	dataTypeStartTimes := dataWork.StringTimeMap{}
	for _, dataType := range test.RandomStringArrayFromArrayWithoutDuplicates(ouraDataWorkPeriodic.DataTypes()) {
		if test.IsOptionalPresent(options...) {
			dataTypeStartTimes[dataType] = pointer.From(test.RandomTimeBeforeNow())
		} else {
			dataTypeStartTimes[dataType] = nil
		}
	}
	return pointer.From(dataTypeStartTimes)
}

func CloneDataTypeStartTimes(datum *dataWork.StringTimeMap) *dataWork.StringTimeMap {
	return dataWorkTest.CloneStringTimeMap(datum)
}

func NewObjectFromDataTypeStartTimes(datum *dataWork.StringTimeMap, format test.ObjectFormat) map[string]any {
	return dataWorkTest.NewObjectFromStringTimeMap(datum, format)
}

func RandomMetadata(options ...test.Option) *ouraDataWorkPeriodic.Metadata {
	return &ouraDataWorkPeriodic.Metadata{
		ProviderSessionMetadata: *providerSessionWorkTest.RandomMetadata(),
		DataTypeNextTokens:      RandomDataTypeNextTokens(options...),
		DataTypeStartTimes:      RandomDataTypeStartTimes(options...),
	}
}

func CloneMetadata(datum *ouraDataWorkPeriodic.Metadata) *ouraDataWorkPeriodic.Metadata {
	if datum == nil {
		return nil
	}
	return &ouraDataWorkPeriodic.Metadata{
		ProviderSessionMetadata: *providerSessionWorkTest.CloneMetadata(&datum.ProviderSessionMetadata),
		DataTypeNextTokens:      CloneDataTypeNextTokens(datum.DataTypeNextTokens),
		DataTypeStartTimes:      CloneDataTypeStartTimes(datum.DataTypeStartTimes),
	}
}

func NewObjectFromMetadata(datum *ouraDataWorkPeriodic.Metadata, format test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	maps.Copy(object, providerSessionWorkTest.NewObjectFromMetadata(&datum.ProviderSessionMetadata, format))
	if datum.DataTypeNextTokens != nil {
		object[ouraDataWorkPeriodic.MetadataKeyDataTypeNextTokens] = NewObjectFromDataTypeNextTokens(datum.DataTypeNextTokens, format)
	}
	if datum.DataTypeStartTimes != nil {
		object[ouraDataWorkPeriodic.MetadataKeyDataTypeStartTimes] = NewObjectFromDataTypeStartTimes(datum.DataTypeStartTimes, format)
	}
	return object
}
