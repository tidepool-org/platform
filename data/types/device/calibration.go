package device

import (
	"reflect"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(unitsField.Tag, UnitsValidator)
	types.GetPlatformValidator().RegisterValidation(valueField.Tag, ValueValidator)
}

type Calibration struct {
	Value *float64 `json:"value" bson:"value" valid:"devicevalue"`
	Units *string  `json:"units" bson:"units" valid:"deviceunits"`
	Base  `bson:",inline"`
}

var (
	unitsField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "units"},
		Tag:        "deviceunits",
		Message:    "Must be one of mg/dl, mmol/l",
		Allowed: types.Allowed{
			"mmol/L": true,
			"mmol/l": true,
			"mg/dL":  true,
			"mg/dl":  true,
		},
	}

	valueField = types.FloatDatumField{
		DatumField:        &types.DatumField{Name: "value"},
		Tag:               "devicevalue",
		Message:           "Must be greater than 0.0",
		AllowedFloatRange: &types.AllowedFloatRange{LowerLimit: 0.0},
	}
)

func (b Base) makeCalibration(datum types.Datum, errs validate.ErrorProcessing) *Calibration {
	Calibration := &Calibration{
		Value: datum.ToFloat64(valueField.Name, errs),
		Units: datum.ToString(unitsField.Name, errs),
		Base:  b,
	}
	types.GetPlatformValidator().Struct(Calibration, errs)
	return Calibration
}

func UnitsValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	units, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = unitsField.Allowed[units]
	return ok
}

func ValueValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	val, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return val > valueField.LowerLimit
}
