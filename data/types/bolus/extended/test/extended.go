package test

import (
	"github.com/tidepool-org/platform/data/types/bolus/extended"
	dataTypesBolusTest "github.com/tidepool-org/platform/data/types/bolus/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewExtended() *extended.Extended {
	datum := extended.New()
	datum.Bolus = *dataTypesBolusTest.RandomBolus()
	datum.SubType = "square"
	datum.Duration = pointer.FromInt(test.RandomIntFromRange(extended.DurationMinimum, extended.DurationMaximum))
	datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, extended.DurationMaximum))
	datum.Extended = pointer.FromFloat64(test.RandomFloat64FromRange(extended.ExtendedMinimum, extended.ExtendedMaximum))
	datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, extended.ExtendedMaximum))
	return datum
}

func CloneExtended(datum *extended.Extended) *extended.Extended {
	if datum == nil {
		return nil
	}
	clone := extended.New()
	clone.Bolus = *dataTypesBolusTest.CloneBolus(&datum.Bolus)
	clone.Duration = pointer.CloneInt(datum.Duration)
	clone.DurationExpected = pointer.CloneInt(datum.DurationExpected)
	clone.Extended = pointer.CloneFloat64(datum.Extended)
	clone.ExtendedExpected = pointer.CloneFloat64(datum.ExtendedExpected)
	return clone
}
