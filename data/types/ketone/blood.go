package ketone

import (
	"fmt"
	"reflect"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(valueField.Tag, ValueValidator)
	types.GetPlatformValidator().RegisterValidation(unitsField.Tag, UnitsValidator)
}

type Blood struct {
	Value      *float64 `json:"value" bson:"value" valid:"ketonebloodvalue"`
	Units      *string  `json:"units" bson:"units" valid:"ketonebloodunits"`
	types.Base `bson:",inline"`
}

const Name = "bloodKetone"

var (
	mmol = "mmol/L"

	unitsField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "units"},
		Tag:        "ketonebloodunits",
		Message:    fmt.Sprintf("Must be %s", mmol),
		Allowed: types.Allowed{
			mmol:     true,
			"mmol/l": true,
		},
	}

	valueField = types.FloatDatumField{
		DatumField:        &types.DatumField{Name: "value"},
		Tag:               "ketonebloodvalue",
		Message:           "Must be greater than 0.0",
		AllowedFloatRange: &types.AllowedFloatRange{LowerLimit: 0.0},
	}

	failureReasons = validate.ErrorReasons{
		valueField.Tag: valueField.Message,
		unitsField.Tag: unitsField.Message,
	}
)

func Build(datum types.Datum, errs validate.ErrorProcessing) *Blood {

	blood := &Blood{
		Value: datum.ToFloat64(valueField.Name, errs),
		Units: datum.ToString(unitsField.Name, errs),
		Base:  types.BuildBase(datum, errs),
	}

	types.GetPlatformValidator().SetErrorReasons(failureReasons).Struct(blood, errs)

	return blood
}

func ValueValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	val, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return val > valueField.LowerLimit
}

func UnitsValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	units, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = unitsField.Allowed[units]
	return ok
}
