package test

import (
	dataTypeInsulin "github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomSimple() *dataTypeInsulin.Simple {
	datum := dataTypeInsulin.NewSimple()
	datum.ActingType = pointer.FromString(test.RandomStringFromArray(dataTypeInsulin.SimpleActingTypes()))
	datum.Brand = pointer.FromString(test.RandomStringFromRange(1, 100))
	datum.Concentration = RandomConcentration()
	return datum
}

func CloneSimple(datum *dataTypeInsulin.Simple) *dataTypeInsulin.Simple {
	if datum == nil {
		return nil
	}
	clone := dataTypeInsulin.NewSimple()
	clone.ActingType = pointer.CloneString(datum.ActingType)
	clone.Brand = pointer.CloneString(datum.Brand)
	clone.Concentration = CloneConcentration(datum.Concentration)
	return clone
}

func NewObjectFromSimple(datum *dataTypeInsulin.Simple, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.ActingType != nil {
		object["actingType"] = test.NewObjectFromString(*datum.ActingType, objectFormat)
	}
	if datum.Brand != nil {
		object["brand"] = test.NewObjectFromString(*datum.Brand, objectFormat)
	}
	if datum.Concentration != nil {
		object["concentration"] = NewObjectFromConcentration(datum.Concentration, objectFormat)
	}
	return object
}
