package test

import (
	"github.com/tidepool-org/platform/data/types/bolus/prescriptor"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewPrescriptor() *prescriptor.Prescriptor {
	datum := prescriptor.NewPrescriptor()
	datum.Prescriptor = pointer.FromString(test.RandomStringFromArray(prescriptor.Presciptors()))
	return datum
}

func ClonePrescriptor(datum *prescriptor.Prescriptor) *prescriptor.Prescriptor {
	if datum == nil {
		return nil
	}
	clone := prescriptor.NewPrescriptor()
	clone.Prescriptor = pointer.CloneString(datum.Prescriptor)
	return clone
}
