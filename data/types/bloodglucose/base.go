package bloodglucose

import (
	"fmt"
	"reflect"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(unitsField.Tag, UnitsValidator)
	types.GetPlatformValidator().RegisterValidation(valueField.Tag, ValueValidator)
}

var (
	mmol = "mmol/L"
	mg   = "mg/dL"

	unitsField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "units"},
		Tag:        "bloodglucoseunits",
		Message:    fmt.Sprintf("Must be one of %s, %s", mmol, mg),
		Allowed: types.Allowed{
			mmol:     true,
			"mmol/l": true,
			mg:       true,
			"mg/dl":  true,
		},
	}

	valueField = types.FloatDatumField{
		DatumField:        &types.DatumField{Name: "value"},
		Tag:               "bloodglucosevalue",
		Message:           "Must be greater than 0.0",
		AllowedFloatRange: &types.AllowedFloatRange{LowerLimit: 0.0},
	}

	failureReasons = validate.ErrorReasons{
		unitsField.Tag: unitsField.Message,
		valueField.Tag: valueField.Message,
		isigField.Tag:  isigField.Message,
	}
)

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

func normalizeUnitName(unitsName *string) *string {
	switch *unitsName {
	case mmol, "mmol/l":
		return &mmol
	case mg, "mg/dl":
		return &mg
	}
	return unitsName
}

func convertMgToMmol(mgValue *float64, units *string) *float64 {

	switch *normalizeUnitName(units) {
	case mg:
		converted := *mgValue / 18.01559
		return &converted
	default:
		return mgValue
	}
}
