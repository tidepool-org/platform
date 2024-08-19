package test

import (
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomUnits(unitsBloodGlucose *string) *dataTypesDosingDecision.Units {
	datum := dataTypesDosingDecision.NewUnits()
	datum.BloodGlucose = unitsBloodGlucose
	datum.Carbohydrate = pointer.FromString(test.RandomStringFromArray(dataTypesDosingDecision.CarbohydrateUnits()))
	datum.Insulin = pointer.FromString(test.RandomStringFromArray(dataTypesDosingDecision.InsulinUnits()))
	return datum
}

func CloneUnits(datum *dataTypesDosingDecision.Units) *dataTypesDosingDecision.Units {
	if datum == nil {
		return nil
	}
	clone := dataTypesDosingDecision.NewUnits()
	clone.BloodGlucose = pointer.CloneString(datum.BloodGlucose)
	clone.Carbohydrate = pointer.CloneString(datum.Carbohydrate)
	clone.Insulin = pointer.CloneString(datum.Insulin)
	return clone
}

func NewObjectFromUnits(datum *dataTypesDosingDecision.Units, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.BloodGlucose != nil {
		object["bg"] = test.NewObjectFromString(*datum.BloodGlucose, objectFormat)
	}
	if datum.Carbohydrate != nil {
		object["carb"] = test.NewObjectFromString(*datum.Carbohydrate, objectFormat)
	}
	if datum.Insulin != nil {
		object["insulin"] = test.NewObjectFromString(*datum.Insulin, objectFormat)
	}
	return object
}
