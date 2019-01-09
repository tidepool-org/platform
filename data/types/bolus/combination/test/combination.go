package test

import (
	"github.com/tidepool-org/platform/data/types/bolus/combination"
	dataTypesBolusTest "github.com/tidepool-org/platform/data/types/bolus/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewCombination() *combination.Combination {
	datum := combination.New()
	datum.Bolus = *dataTypesBolusTest.NewBolus()
	datum.SubType = "dual/square"
	datum.Duration = pointer.FromInt(test.RandomIntFromRange(combination.DurationMinimum, combination.DurationMaximum))
	datum.Extended = pointer.FromFloat64(test.RandomFloat64FromRange(combination.ExtendedMinimum, combination.ExtendedMaximum))
	datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, combination.DurationMaximum))
	datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, combination.ExtendedMaximum))
	datum.Normal = pointer.FromFloat64(test.RandomFloat64FromRange(combination.NormalMinimum, combination.NormalMaximum))
	datum.NormalExpected = nil
	return datum
}

func CloneCombination(datum *combination.Combination) *combination.Combination {
	if datum == nil {
		return nil
	}
	clone := combination.New()
	clone.Duration = pointer.CloneInt(datum.Duration)
	clone.DurationExpected = pointer.CloneInt(datum.DurationExpected)
	clone.Extended = pointer.CloneFloat64(datum.Extended)
	clone.ExtendedExpected = pointer.CloneFloat64(datum.ExtendedExpected)
	clone.Bolus = *dataTypesBolusTest.CloneBolus(&datum.Bolus)
	clone.Normal = pointer.CloneFloat64(datum.Normal)
	clone.NormalExpected = pointer.CloneFloat64(datum.NormalExpected)
	return clone
}
