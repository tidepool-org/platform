package basal

import (
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/validate"
)

type Suspend struct {
	Suppressed *Suppressed `json:"suppressed,omitempty" bson:"suppressed,omitempty" valid:"-"`
	Base       `bson:",inline"`
}

func (b Base) makeSuspend(datum types.Datum, errs validate.ErrorProcessing) *Suspend {

	var suppressed *Suppressed
	suppressedDatum, ok := datum["suppressed"].(map[string]interface{})
	if ok {
		suppressed = makeSuppressed(suppressedDatum, errs)
	}

	suspend := &Suspend{
		Suppressed: suppressed,
		Base:       b,
	}
	types.GetPlatformValidator().SetErrorReasons(failureReasons).Struct(suspend, errs)
	return suspend

}
