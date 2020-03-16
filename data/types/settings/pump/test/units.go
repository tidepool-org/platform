package test

import (
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewUnits(unitsBloodGlucose *string) *pump.Units {
	datum := pump.NewUnits()
	datum.BloodGlucose = unitsBloodGlucose
	datum.Carbohydrate = pointer.FromString(test.RandomStringFromArray(pump.Carbohydrates()))
	return datum
}

func CloneUnits(datum *pump.Units) *pump.Units {
	if datum == nil {
		return nil
	}
	clone := pump.NewUnits()
	clone.BloodGlucose = pointer.CloneString(datum.BloodGlucose)
	clone.Carbohydrate = pointer.CloneString(datum.Carbohydrate)
	return clone
}
