package test

import (
	"maps"
	"time"

	dataWork "github.com/tidepool-org/platform/data/work"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomStringTimeMapReference() string {
	return test.RandomStringFromRange(1, dataWork.StringTimeMapReferenceLengthMaximum)
}

func RandomStringTimeMapValue() time.Time {
	return test.RandomTime()
}

func RandomStringTimeMap() *dataWork.StringTimeMap {
	datum := dataWork.StringTimeMap{}
	for range test.RandomIntFromRange(1, 3) {
		datum[RandomStringTimeMapReference()] = pointer.From(RandomStringTimeMapValue())
	}
	return pointer.From(datum)
}

func CloneStringTimeMap(datum *dataWork.StringTimeMap) *dataWork.StringTimeMap {
	if datum == nil {
		return nil
	}
	return pointer.From(maps.Clone(*datum))
}

func NewObjectFromStringTimeMap(datum *dataWork.StringTimeMap, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	for reference, value := range *datum {
		if value != nil {
			object[reference] = test.NewObjectFromTime(*value, objectFormat)
		} else {
			object[reference] = nil
		}
	}
	return object
}
