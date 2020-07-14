package test

import (
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewBasalTemporary() *pump.BasalTemporary {
	datum := pump.NewBasalTemporary()
	datum.Type = pointer.FromString(test.RandomStringFromArray(pump.BasalTemporaryTypes()))
	return datum
}

func CloneBasalTemporary(datum *pump.BasalTemporary) *pump.BasalTemporary {
	if datum == nil {
		return nil
	}
	clone := pump.NewBasalTemporary()
	clone.Type = pointer.CloneString(datum.Type)
	return clone
}
