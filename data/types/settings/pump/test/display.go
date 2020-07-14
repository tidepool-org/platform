package test

import "github.com/tidepool-org/platform/data/types/settings/pump"

func NewDisplay() *pump.Display {
	datum := pump.NewDisplay()
	datum.BloodGlucose = NewDisplayBloodGlucose()
	return datum
}

func CloneDisplay(datum *pump.Display) *pump.Display {
	if datum == nil {
		return nil
	}
	clone := pump.NewDisplay()
	clone.BloodGlucose = CloneDisplayBloodGlucose(datum.BloodGlucose)
	return clone
}
