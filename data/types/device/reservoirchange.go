package device

import (
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

type ReservoirChange struct {
	Status *string `json:"status" bson:"status" valid:"devicestatus"`
	Base   `bson:",inline"`
}

func (b Base) makeReservoirChange(datum types.Datum, errs validate.ErrorProcessing) *ReservoirChange {
	reservoirChange := &ReservoirChange{
		Status: datum.ToString(statusField.Name, errs),
		Base:   b,
	}
	types.GetPlatformValidator().Struct(reservoirChange, errs)
	return reservoirChange
}
