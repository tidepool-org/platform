package test

import (
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewBolusCalculatorInsulin() *pump.BolusCalculatorInsulin {
	units := pointer.FromString(test.RandomStringFromArray(pump.BolusCalculatorInsulinUnits()))
	datum := pump.NewBolusCalculatorInsulin()
	datum.Duration = pointer.FromFloat64(test.RandomFloat64FromRange(pump.BolusCalculatorInsulinDurationRangeForUnits(units)))
	datum.Units = units
	return datum
}

func CloneBolusCalculatorInsulin(datum *pump.BolusCalculatorInsulin) *pump.BolusCalculatorInsulin {
	if datum == nil {
		return nil
	}
	clone := pump.NewBolusCalculatorInsulin()
	clone.Duration = pointer.CloneFloat64(datum.Duration)
	clone.Units = pointer.CloneString(datum.Units)
	return clone
}
