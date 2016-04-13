package bolus

import (
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

type DualSquare struct {
	Normal   *float64 `json:"normal,omitempty" bson:"normal,omitempty" valid:"omitempty,bolusnormal"`
	Extended *float64 `json:"extended,omitempty" bson:"extended,omitempty" valid:"omitempty,bolusextended"`
	Duration *int     `json:"duration,omitempty" bson:"duration,omitempty" valid:"omitempty,bolusduration"`
	Base     `bson:",inline"`
}

func (b Base) makeDualSquare(datum types.Datum, errs validate.ErrorProcessing) *DualSquare {
	dualSquare := &DualSquare{
		Duration: datum.ToInt(durationField.Name, errs),
		Extended: datum.ToFloat64(extendedField.Name, errs),
		Normal:   datum.ToFloat64(normalField.Name, errs),
		Base:     b,
	}
	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(dualSquare, errs)
	return dualSquare
}
