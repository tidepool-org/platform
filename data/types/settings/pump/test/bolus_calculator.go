package test

import (
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewBolusCalculator() *pump.BolusCalculator {
	datum := pump.NewBolusCalculator()
	datum.Enabled = pointer.FromBool(test.RandomBool())
	datum.Insulin = NewBolusCalculatorInsulin()
	return datum
}

func CloneBolusCalculator(datum *pump.BolusCalculator) *pump.BolusCalculator {
	if datum == nil {
		return nil
	}
	clone := pump.NewBolusCalculator()
	clone.Enabled = pointer.CloneBool(datum.Enabled)
	clone.Insulin = CloneBolusCalculatorInsulin(datum.Insulin)
	return clone
}
