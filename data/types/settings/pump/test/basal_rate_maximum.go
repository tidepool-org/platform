package test

import (
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewBasalRateMaximum() *pump.BasalRateMaximum {
	datum := pump.NewBasalRateMaximum()
	datum.Units = pointer.FromString(test.RandomStringFromArray(pump.BasalRateMaximumUnits()))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(pump.BasalRateMaximumValueRangeForUnits(datum.Units)))
	return datum
}

func CloneBasalRateMaximum(datum *pump.BasalRateMaximum) *pump.BasalRateMaximum {
	if datum == nil {
		return nil
	}
	clone := pump.NewBasalRateMaximum()
	clone.Units = pointer.CloneString(datum.Units)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}
