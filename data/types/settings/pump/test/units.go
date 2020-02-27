package test

import (
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
)

func CloneUnits(datum *pump.Units) *pump.Units {
	if datum == nil {
		return nil
	}
	clone := pump.NewUnits()
	clone.BloodGlucose = pointer.CloneString(datum.BloodGlucose)
	clone.Carbohydrate = pointer.CloneString(datum.Carbohydrate)
	return clone
}
