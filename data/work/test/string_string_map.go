package test

import (
	"maps"

	dataWork "github.com/tidepool-org/platform/data/work"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomStringStringMapReference() string {
	return test.RandomStringFromRange(1, dataWork.StringStringMapReferenceLengthMaximum)
}

func RandomStringStringMapValue() string {
	return test.RandomStringFromRange(1, dataWork.StringStringMapValueLengthMaximum)
}

func RandomStringStringMap() *dataWork.StringStringMap {
	datum := dataWork.StringStringMap{}
	for range test.RandomIntFromRange(1, 3) {
		datum[RandomStringStringMapReference()] = pointer.From(RandomStringStringMapValue())
	}
	return pointer.From(datum)
}

func CloneStringStringMap(datum *dataWork.StringStringMap) *dataWork.StringStringMap {
	if datum == nil {
		return nil
	}
	return pointer.From(maps.Clone(*datum))
}

func NewObjectFromStringStringMap(datum *dataWork.StringStringMap, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	for reference, value := range *datum {
		if value != nil {
			object[reference] = test.NewObjectFromString(*value, objectFormat)
		} else {
			object[reference] = nil
		}
	}
	return object
}
