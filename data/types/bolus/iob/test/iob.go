package test

import (
	"github.com/tidepool-org/platform/data/types/bolus/iob"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewIob() *iob.Iob {
	datum := iob.NewIob()
	datum.InsulinOnBoard = pointer.FromFloat64(test.RandomFloat64FromRange(iob.InsulinOnBoardMinimum, iob.InsulinOnBoardMaximum))
	return datum
}

func CloneIob(datum *iob.Iob) *iob.Iob {
	if datum == nil {
		return nil
	}
	clone := iob.NewIob()
	clone.InsulinOnBoard = pointer.CloneFloat64(datum.InsulinOnBoard)
	return clone
}
