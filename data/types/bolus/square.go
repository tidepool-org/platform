package bolus

import (
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

type Square struct {
	Extended *float64 `json:"extended,omitempty" bson:"extended,omitempty" valid:"omitempty,bolusextended"`
	Duration *int     `json:"duration,omitempty" bson:"duration,omitempty" valid:"omitempty,bolusduration"`
	Base     `bson:",inline"`
}

func (b Base) makeSquare(datum types.Datum, errs validate.ErrorProcessing) *Square {
	square := &Square{
		Duration: datum.ToInt(durationField.Name, errs),
		Extended: datum.ToFloat64(extendedField.Name, errs),
		Base:     b,
	}
	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(square, errs)
	return square
}
