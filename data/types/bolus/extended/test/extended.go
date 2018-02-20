package test

import (
	"github.com/tidepool-org/platform/data/types/bolus/extended"
	testDataTypesBolus "github.com/tidepool-org/platform/data/types/bolus/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewExtended() *extended.Extended {
	datum := extended.New()
	datum.Bolus = *testDataTypesBolus.NewBolus()
	datum.SubType = "square"
	datum.Duration = pointer.Int(test.RandomIntFromRange(extended.DurationMinimum, extended.DurationMaximum))
	datum.DurationExpected = pointer.Int(test.RandomIntFromRange(*datum.Duration, extended.DurationMaximum))
	datum.Extended = pointer.Float64(test.RandomFloat64FromRange(extended.ExtendedMinimum, extended.ExtendedMaximum))
	datum.ExtendedExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Extended, extended.ExtendedMaximum))
	return datum
}

func CloneExtended(datum *extended.Extended) *extended.Extended {
	if datum == nil {
		return nil
	}
	clone := extended.New()
	clone.Bolus = *testDataTypesBolus.CloneBolus(&datum.Bolus)
	clone.Duration = test.CloneInt(datum.Duration)
	clone.DurationExpected = test.CloneInt(datum.DurationExpected)
	clone.Extended = test.CloneFloat64(datum.Extended)
	clone.ExtendedExpected = test.CloneFloat64(datum.ExtendedExpected)
	return clone
}
