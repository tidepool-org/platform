package test

import (
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewBolusExtended() *pump.BolusExtended {
	datum := pump.NewBolusExtended()
	datum.Enabled = pointer.FromBool(test.RandomBool())
	return datum
}

func CloneBolusExtended(datum *pump.BolusExtended) *pump.BolusExtended {
	if datum == nil {
		return nil
	}
	clone := pump.NewBolusExtended()
	clone.Enabled = pointer.CloneBool(datum.Enabled)
	return clone
}
