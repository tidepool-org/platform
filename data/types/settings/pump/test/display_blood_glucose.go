package test

import (
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewDisplayBloodGlucose() *pump.DisplayBloodGlucose {
	datum := pump.NewDisplayBloodGlucose()
	datum.Units = pointer.FromString(test.RandomStringFromArray(pump.DisplayBloodGlucoseUnits()))
	return datum
}

func CloneDisplayBloodGlucose(datum *pump.DisplayBloodGlucose) *pump.DisplayBloodGlucose {
	if datum == nil {
		return nil
	}
	clone := pump.NewDisplayBloodGlucose()
	clone.Units = pointer.CloneString(datum.Units)
	return clone
}
