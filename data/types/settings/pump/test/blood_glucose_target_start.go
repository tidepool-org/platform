package test

import (
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewBloodGlucoseTargetStartTest(units *string, startMinimum int) *pump.BloodGlucoseTargetStart {
	datum := pump.NewBloodGlucoseTargetStart()
	datum.Target = *dataBloodGlucoseTest.NewTarget(units)
	if startMinimum == pump.BloodGlucoseTargetStartStartMinimum {
		datum.Start = pointer.FromInt(pump.BloodGlucoseTargetStartStartMinimum)
	} else {
		datum.Start = pointer.FromInt(test.RandomIntFromRange(startMinimum, pump.BloodGlucoseTargetStartStartMaximum))
	}
	return datum
}

func NewBloodGlucoseTargetStartArrayTest(units *string) *pump.BloodGlucoseTargetStartArray {
	datum := pump.NewBloodGlucoseTargetStartArray()
	*datum = append(*datum, NewBloodGlucoseTargetStartTest(units, pump.BloodGlucoseTargetStartStartMinimum))
	*datum = append(*datum, NewBloodGlucoseTargetStartTest(units, *datum.Last().Start+1))
	*datum = append(*datum, NewBloodGlucoseTargetStartTest(units, *datum.Last().Start+1))
	return datum
}
