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
