package bolus

import (
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

type Combo struct {
	Normal   *float64 `json:"normal" bson:"normal" valid:"bolusnormal"`
	Extended *float64 `json:"extended" bson:"extended" valid:"bolusextended"`
	Duration *int     `json:"duration" bson:"duration" valid:"bolusduration"`
	Base     `bson:",inline"`
}

func (b Base) makeCombo(datum types.Datum, errs validate.ErrorProcessing) *Combo {
	combo := &Combo{
		Duration: datum.ToInt(durationField.Name, errs),
		Extended: datum.ToFloat64(extendedField.Name, errs),
		Normal:   datum.ToFloat64(normalField.Name, errs),
		Base:     b,
	}
	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(combo, errs)
	return combo
}
