package test

import (
	"maps"

	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomMetadataMap() map[string]any {
	datum := map[string]any{}
	for range test.RandomIntFromRange(1, 3) {
		datum[RandomMetadataKey()] = RandomMetadataValue()
	}
	return datum
}

func RandomOptionalMetadataMap(options ...test.Option) map[string]any {
	if test.IsConditionallyTrue(options...) {
		return RandomMetadataMap()
	} else {
		return nil
	}
}

func CloneMetadataMap(datum map[string]any) map[string]any {
	return maps.Clone(datum)
}

func NewObjectFromMetadataMap(datum map[string]any, objectFormat test.ObjectFormat) map[string]any {
	return maps.Clone(datum)
}

func PointerFromMetadataMap(datum map[string]any) *map[string]any {
	if datum == nil {
		return nil
	}
	return &datum
}

func RandomMetadataMapPointer() *map[string]any {
	return pointer.From(RandomMetadataMap())
}

func CloneMetadataMapPointer(datum *map[string]any) *map[string]any {
	if datum == nil {
		return nil
	}
	return pointer.From(CloneMetadataMap(*datum))
}

func NewObjectFromMetadataMapPointer(datum *map[string]any, objectFormat test.ObjectFormat) *map[string]any {
	if datum == nil {
		return nil
	}
	return pointer.From(NewObjectFromMetadataMap(*datum, objectFormat))
}

func RandomMetadata() *metadata.Metadata {
	return pointer.From(metadata.Metadata(RandomMetadataMap()))
}

func CloneMetadata(datum *metadata.Metadata) *metadata.Metadata {
	if datum == nil {
		return nil
	}
	return pointer.From(metadata.Metadata(CloneMetadataMap(*datum)))
}

func NewObjectFromMetadata(datum *metadata.Metadata, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	return NewObjectFromMetadataMap(*datum, objectFormat)
}

func RandomMetadataKey() string {
	return test.RandomStringFromCharset(test.CharsetAlphaNumeric)
}

func RandomMetadataValue() any {
	return test.RandomString()
}

func RandomMetadataArray() *metadata.MetadataArray {
	datumArray := metadata.NewMetadataArray()
	for range test.RandomIntFromRange(1, 3) {
		*datumArray = append(*datumArray, RandomMetadata())
	}
	return datumArray
}

func CloneMetadataArray(datumArray *metadata.MetadataArray) *metadata.MetadataArray {
	if datumArray == nil {
		return nil
	}
	cloneArray := metadata.NewMetadataArray()
	for _, datum := range *datumArray {
		*cloneArray = append(*cloneArray, CloneMetadata(datum))
	}
	return cloneArray
}

func NewArrayFromMetadataArray(datumArray *metadata.MetadataArray, objectFormat test.ObjectFormat) []any {
	if datumArray == nil {
		return nil
	}
	array := []any{}
	for _, datum := range *datumArray {
		array = append(array, NewObjectFromMetadata(datum, objectFormat))
	}
	return array
}
