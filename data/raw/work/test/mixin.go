package test

import (
	dataRawTest "github.com/tidepool-org/platform/data/raw/test"
	dataRawWork "github.com/tidepool-org/platform/data/raw/work"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomMetadata(options ...test.Option) *dataRawWork.Metadata {
	return &dataRawWork.Metadata{
		DataRawID: test.RandomOptional(dataRawTest.RandomDataRawID, options...),
	}
}

func CloneMetadata(datum *dataRawWork.Metadata) *dataRawWork.Metadata {
	if datum == nil {
		return nil
	}
	return &dataRawWork.Metadata{
		DataRawID: pointer.Clone(datum.DataRawID),
	}
}

func NewObjectFromMetadata(datum *dataRawWork.Metadata, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.DataRawID != nil {
		object[dataRawWork.MetadataKeyDataRawID] = test.NewObjectFromString(*datum.DataRawID, objectFormat)
	}
	return object
}
