package test

import (
	"github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewPrescriptor() *common.Prescriptor {
	datum := common.NewPrescriptor()
	datum.Prescriptor = pointer.FromString(test.RandomStringFromArray(common.Presciptors()))
	return datum
}

func ClonePrescriptor(datum *common.Prescriptor) *common.Prescriptor {
	if datum == nil {
		return nil
	}
	clone := common.NewPrescriptor()
	clone.Prescriptor = pointer.CloneString(datum.Prescriptor)
	return clone
}
