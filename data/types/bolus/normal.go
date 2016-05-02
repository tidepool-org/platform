package bolus

import (
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

type Normal struct {
	Normal *float64 `json:"normal" bson:"normal" valid:"bolusnormal"`
	Base   `bson:",inline"`
}

func (b Base) makeNormal(datum types.Datum, errs validate.ErrorProcessing) *Normal {
	normal := &Normal{
		Normal: datum.ToFloat64(normalField.Name, errs),
		Base:   b,
	}
	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(normal, errs)
	return normal
}
