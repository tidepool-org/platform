package basal

import (
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/validate"
)

type Suspend struct {
	Suppressed *SuppressedBasal `json:"suppressed,omitempty" bson:"suppressed,omitempty" valid:"omitempty,required"`
	Base       `bson:",inline"`
}

func (b Base) makeSuspend(datum types.Datum, errs validate.ErrorProcessing) *Suspend {

	suspend := &Suspend{
		Suppressed: makeSuppressed(datum["suppressed"].(map[string]interface{}), errs),
		Base:       b,
	}
	types.GetPlatformValidator().SetErrorReasons(failureReasons).Struct(suspend, errs)
	return suspend
}
