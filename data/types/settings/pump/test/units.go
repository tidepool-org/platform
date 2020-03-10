package test

import (
	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomUnits(unitsBloodGlucose *string) *dataTypesSettingsPump.Units {
	datum := dataTypesSettingsPump.NewUnits()
	datum.BloodGlucose = unitsBloodGlucose
	datum.Carbohydrate = pointer.FromString(test.RandomStringFromArray(dataTypesSettingsPump.Carbohydrates()))
	return datum
}

func CloneUnits(datum *dataTypesSettingsPump.Units) *dataTypesSettingsPump.Units {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsPump.NewUnits()
	clone.BloodGlucose = pointer.CloneString(datum.BloodGlucose)
	clone.Carbohydrate = pointer.CloneString(datum.Carbohydrate)
	return clone
}
