package test

import (
	dataSetWork "github.com/tidepool-org/platform/data/set/work"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomMetadata(options ...test.Option) *dataSetWork.Metadata {
	return &dataSetWork.Metadata{
		DataSetID: test.RandomOptional(dataTest.RandomDataSetID, options...),
	}
}

func CloneMetadata(datum *dataSetWork.Metadata) *dataSetWork.Metadata {
	if datum == nil {
		return nil
	}
	return &dataSetWork.Metadata{
		DataSetID: pointer.CloneString(datum.DataSetID),
	}
}

func NewObjectFromMetadata(datum *dataSetWork.Metadata, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.DataSetID != nil {
		object[dataSetWork.MetadataKeyDataSetID] = test.NewObjectFromString(*datum.DataSetID, objectFormat)
	}
	return object
}
