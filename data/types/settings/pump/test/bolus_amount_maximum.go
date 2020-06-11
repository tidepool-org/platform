package test

import (
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewBolusAmountMaximum() *pump.BolusAmountMaximum {
	datum := pump.NewBolusAmountMaximum()
	datum.Units = pointer.FromString(test.RandomStringFromArray(pump.BolusAmountMaximumUnits()))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(pump.BolusAmountMaximumValueRangeForUnits(datum.Units)))
	return datum
}

func CloneBolusAmountMaximum(datum *pump.BolusAmountMaximum) *pump.BolusAmountMaximum {
	if datum == nil {
		return nil
	}
	clone := pump.NewBolusAmountMaximum()
	clone.Units = pointer.CloneString(datum.Units)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}
