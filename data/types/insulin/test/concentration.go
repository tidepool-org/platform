package test

import (
	dataTypeInsulin "github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomConcentration() *dataTypeInsulin.Concentration {
	datum := dataTypeInsulin.NewConcentration()
	datum.Units = pointer.FromString(test.RandomStringFromArray(dataTypeInsulin.ConcentrationUnits()))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypeInsulin.ConcentrationValueRangeForUnits(datum.Units)))
	return datum
}

func CloneConcentration(datum *dataTypeInsulin.Concentration) *dataTypeInsulin.Concentration {
	if datum == nil {
		return nil
	}
	clone := dataTypeInsulin.NewConcentration()
	clone.Units = pointer.CloneString(datum.Units)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}

func NewObjectFromConcentration(datum *dataTypeInsulin.Concentration, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Units != nil {
		object["units"] = test.NewObjectFromString(*datum.Units, objectFormat)
	}
	if datum.Value != nil {
		object["value"] = test.NewObjectFromFloat64(*datum.Value, objectFormat)
	}
	return object
}
