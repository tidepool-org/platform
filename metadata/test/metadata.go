package test

import (
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/test"
)

func RandomMetadataMap() map[string]any {
	return RandomMetadata().AsMap()
}

func CloneMetadataMap(datum map[string]any) map[string]any {
	if datum == nil {
		return nil
	}
	clone := map[string]any{}
	for key, value := range datum {
		clone[key] = value
	}
	return clone
}

func NewObjectFromMetadataMap(datum map[string]any, objectFormat test.ObjectFormat) map[string]interface{} {
	return datum
}

func RandomMetadata() *metadata.Metadata {
	datum := metadata.NewMetadata()
	for index := test.RandomIntFromRange(1, 3); index > 0; index-- {
		(*datum)[RandomMetadataKey()] = RandomMetadataValue()
	}
	return datum
}

func CloneMetadata(datum *metadata.Metadata) *metadata.Metadata {
	if datum == nil {
		return nil
	}
	clone := metadata.NewMetadata()
	for key, value := range *datum {
		(*clone)[key] = value
	}
	return clone
}

func NewObjectFromMetadata(datum *metadata.Metadata, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	for key, value := range *datum {
		object[key] = value
	}
	return object
}

func RandomMetadataKey() string {
	return test.RandomStringFromCharset(test.CharsetAlphaNumeric)
}

func RandomMetadataValue() interface{} {
	return test.RandomString()
}

func RandomMetadataArray() *metadata.MetadataArray {
	datumArray := metadata.NewMetadataArray()
	for index := test.RandomIntFromRange(1, 3); index > 0; index-- {
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

func NewArrayFromMetadataArray(datumArray *metadata.MetadataArray, objectFormat test.ObjectFormat) []interface{} {
	if datumArray == nil {
		return nil
	}
	array := []interface{}{}
	for _, datum := range *datumArray {
		array = append(array, NewObjectFromMetadata(datum, objectFormat))
	}
	return array
}
