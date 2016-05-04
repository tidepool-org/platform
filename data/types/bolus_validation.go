package types

import (
	"reflect"

	validator "gopkg.in/bluesuncorp/validator.v8"
)

func init() {
	GetPlatformValidator().RegisterValidation(BolusSubTypeField.Tag, BolusSubTypeValidator)
}

var BolusSubTypeField = DatumFieldInformation{
	DatumField: &DatumField{Name: "subType"},
	Tag:        "bolussubtype",
	Message:    "Must be one of normal, square, dual/square",
	Allowed:    Allowed{"normal": true, "square": true, "dual/square": true},
}

func BolusSubTypeValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	subType, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = BolusSubTypeField.Allowed[subType]
	return ok
}
