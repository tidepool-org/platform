package ketone

import (
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

type Blood struct {
	Value      *float64 `json:"value" bson:"value" valid:"bloodglucosevalue"`
	Units      *string  `json:"units" bson:"units" valid:"mmolunits"`
	types.Base `bson:",inline"`
}

const Name = "bloodKetone"

var (
	failureReasons = validate.FailureReasons{
		"Value": validate.VaidationInfo{FieldName: types.BloodGlucoseValueField.Name, Message: types.BloodGlucoseValueField.Message},
		"Units": validate.VaidationInfo{FieldName: types.MmolUnitsField.Name, Message: types.MmolUnitsField.Message},
	}
)

func Build(datum types.Datum, errs validate.ErrorProcessing) *Blood {

	blood := &Blood{
		Value: datum.ToFloat64(types.BloodGlucoseValueField.Name, errs),
		Units: datum.ToString(types.MmolUnitsField.Name, errs),
		Base:  types.BuildBase(datum, errs),
	}

	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(blood, errs)

	return blood
}
