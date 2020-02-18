package pump

import (
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewBloodGlucoseTargetStartTest(units *string, startMinimum int) *BloodGlucoseTargetStart {
	datum := NewBloodGlucoseTargetStart()
	datum.Target = *dataBloodGlucoseTest.NewTarget(units)
	if startMinimum == BloodGlucoseTargetStartStartMinimum {
		datum.Start = pointer.FromInt(BloodGlucoseTargetStartStartMinimum)
	} else {
		datum.Start = pointer.FromInt(test.RandomIntFromRange(startMinimum, BloodGlucoseTargetStartStartMaximum))
	}
	return datum
}

func NewBloodGlucoseTargetStartArrayTest(units *string) *BloodGlucoseTargetStartArray {
	datum := NewBloodGlucoseTargetStartArray()
	*datum = append(*datum, NewBloodGlucoseTargetStartTest(units, BloodGlucoseTargetStartStartMinimum))
	*datum = append(*datum, NewBloodGlucoseTargetStartTest(units, *datum.Last().Start+1))
	*datum = append(*datum, NewBloodGlucoseTargetStartTest(units, *datum.Last().Start+1))
	return datum
}
