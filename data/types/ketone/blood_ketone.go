package ketone

import (
	"reflect"

	validator "gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(valueField.Tag, ValueValidator)
}

type Blood struct {
	Value      *float64 `json:"value" bson:"value" valid:"ketonevalue"`
	Units      *string  `json:"units" bson:"units" valid:"mmolunits"`
	types.Base `bson:",inline"`
}

const Name = "bloodKetone"

var (
	valueField = types.FloatDatumField{
		DatumField:        &types.DatumField{Name: "value"},
		Tag:               "ketonevalue",
		Message:           "Needs to be in the range of >= 0.0 and <= 10.0",
		AllowedFloatRange: &types.AllowedFloatRange{LowerLimit: 0.0, UpperLimit: 10.0},
	}

	failureReasons = validate.FailureReasons{
		"Value": validate.ValidationInfo{FieldName: valueField.Name, Message: valueField.Message},
		"Units": validate.ValidationInfo{FieldName: types.MmolUnitsField.Name, Message: types.MmolUnitsField.Message},
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

func ValueValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	val, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return val >= valueField.LowerLimit && val <= valueField.UpperLimit
}
