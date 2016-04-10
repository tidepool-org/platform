package ketone

import (
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/validate"
)

type Blood struct {
	Value      *float64 `json:"value" bson:"value" valid:"bloodvalue"`
	Units      *string  `json:"units" bson:"units" valid:"mmolunits"`
	types.Base `bson:",inline"`
}

const Name = "bloodKetone"

var (
	failureReasons = validate.ErrorReasons{
		types.BloodValueField.Tag: types.BloodValueField.Message,
		types.MmolUnitsField.Tag:  types.MmolUnitsField.Message,
	}
)

func Build(datum types.Datum, errs validate.ErrorProcessing) *Blood {

	blood := &Blood{
		Value: datum.ToFloat64(types.BloodValueField.Name, errs),
		Units: datum.ToString(types.MmolUnitsField.Name, errs),
		Base:  types.BuildBase(datum, errs),
	}

	types.GetPlatformValidator().SetErrorReasons(failureReasons).Struct(blood, errs)

	return blood
}
