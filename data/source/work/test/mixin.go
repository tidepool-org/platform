package test

import (
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	dataSourceWork "github.com/tidepool-org/platform/data/source/work"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomMetadata(options ...test.Option) *dataSourceWork.Metadata {
	return &dataSourceWork.Metadata{
		DataSourceID: test.RandomOptional(dataSourceTest.RandomDataSourceID, options...),
	}
}

func CloneMetadata(datum *dataSourceWork.Metadata) *dataSourceWork.Metadata {
	if datum == nil {
		return nil
	}
	return &dataSourceWork.Metadata{
		DataSourceID: pointer.CloneString(datum.DataSourceID),
	}
}

func NewObjectFromMetadata(datum *dataSourceWork.Metadata, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.DataSourceID != nil {
		object[dataSourceWork.MetadataKeyDataSourceID] = test.NewObjectFromString(*datum.DataSourceID, objectFormat)
	}
	return object
}
