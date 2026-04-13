package test

import (
	dataWork "github.com/tidepool-org/platform/data/work"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomIngestionOffset() int {
	return test.RandomIntFromRange(0, test.RandomIntMaximum())
}

func RandomIngestionOffsetMetadata(options ...test.Option) *dataWork.IngestionOffsetMetadata {
	return &dataWork.IngestionOffsetMetadata{
		IngestionOffset: test.RandomOptional(RandomIngestionOffset, options...),
	}
}

func CloneIngestionOffsetMetadata(datum *dataWork.IngestionOffsetMetadata) *dataWork.IngestionOffsetMetadata {
	if datum == nil {
		return nil
	}
	return &dataWork.IngestionOffsetMetadata{
		IngestionOffset: pointer.Clone(datum.IngestionOffset),
	}
}

func NewObjectFromIngestionOffsetMetadata(datum *dataWork.IngestionOffsetMetadata, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.IngestionOffset != nil {
		object[dataWork.MetadataKeyIngestionOffset] = test.NewObjectFromInt(*datum.IngestionOffset, objectFormat)
	}
	return object
}
