package test

import (
	"github.com/tidepool-org/platform/data/types/bolus/combination"
	testDataTypesBolus "github.com/tidepool-org/platform/data/types/bolus/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewCombination() *combination.Combination {
	datum := combination.New()
	datum.Bolus = *testDataTypesBolus.NewBolus()
	datum.SubType = "dual/square"
	datum.Duration = pointer.Int(test.RandomIntFromRange(combination.DurationMinimum, combination.DurationMaximum))
	datum.Extended = pointer.Float64(test.RandomFloat64FromRange(combination.ExtendedMinimum, combination.ExtendedMaximum))
	datum.DurationExpected = pointer.Int(test.RandomIntFromRange(*datum.Duration, combination.DurationMaximum))
	datum.ExtendedExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Extended, combination.ExtendedMaximum))
	datum.Normal = pointer.Float64(test.RandomFloat64FromRange(combination.NormalMinimum, combination.NormalMaximum))
	datum.NormalExpected = nil
	return datum
}

func CloneCombination(datum *combination.Combination) *combination.Combination {
	if datum == nil {
		return nil
	}
	clone := combination.New()
	clone.Duration = test.CloneInt(datum.Duration)
	clone.DurationExpected = test.CloneInt(datum.DurationExpected)
	clone.Extended = test.CloneFloat64(datum.Extended)
	clone.ExtendedExpected = test.CloneFloat64(datum.ExtendedExpected)
	clone.Bolus = *testDataTypesBolus.CloneBolus(&datum.Bolus)
	clone.Normal = test.CloneFloat64(datum.Normal)
	clone.NormalExpected = test.CloneFloat64(datum.NormalExpected)
	return clone
}
