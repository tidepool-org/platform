package test

import (
	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomUnits(unitsBloodGlucose *string) *dataTypesSettingsPump.Units {
	datum := dataTypesSettingsPump.NewUnits()
	datum.BloodGlucose = unitsBloodGlucose
	datum.Carbohydrate = pointer.FromString(test.RandomStringFromArray(dataTypesSettingsPump.CarbohydrateUnits()))
	datum.Insulin = pointer.FromString(test.RandomStringFromArray(dataTypesSettingsPump.InsulinUnits()))
	return datum
}

func CloneUnits(datum *dataTypesSettingsPump.Units) *dataTypesSettingsPump.Units {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsPump.NewUnits()
	clone.BloodGlucose = pointer.CloneString(datum.BloodGlucose)
	clone.Carbohydrate = pointer.CloneString(datum.Carbohydrate)
	clone.Insulin = pointer.CloneString(datum.Insulin)
	return clone
}

func NewObjectFromUnits(datum *dataTypesSettingsPump.Units, objectFormat test.ObjectFormat) map[string]interface{} {
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
