package bolus

import (
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

type Extended struct {
	Extended *float64 `json:"extended" bson:"extended" valid:"bolusextended"`
	Duration *int     `json:"duration" bson:"duration" valid:"bolusduration"`
	Base     `bson:",inline"`
}

func (b Base) makeExtended(datum types.Datum, errs validate.ErrorProcessing) *Extended {
	extended := &Extended{
		Duration: datum.ToInt(durationField.Name, errs),
		Extended: datum.ToFloat64(extendedField.Name, errs),
		Base:     b,
	}
	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(extended, errs)
	return extended
}
