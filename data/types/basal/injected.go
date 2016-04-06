package basal

import (
	"reflect"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(insulinField.Tag, InsulinValidator)
	types.GetPlatformValidator().RegisterValidation(valueField.Tag, ValueValidator)
}

type Injected struct {
	Insulin *string `json:"insulin,omitempty" bson:"insulin,omitempty" valid:"omitempty,basalinsulin"`
	Value   *int    `json:"value,omitempty" bson:"value,omitempty" valid:"omitempty,basalvalue"`
	Base    `bson:",inline"`
}

var (
	insulinField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "insulin"},
		Tag:        "basalinsulin",
		Message:    "Must be one of levemir, lantus",
		Allowed:    types.Allowed{"levemir": true, "lantus": true},
	}

	valueField = types.IntDatumField{
		DatumField:      &types.DatumField{Name: "value"},
		Tag:             "basalvalue",
		Message:         "Must be greater than 0",
		AllowedIntRange: &types.AllowedIntRange{LowerLimit: 0},
	}
)

func (b Base) makeInjected(datum types.Datum, errs validate.ErrorProcessing) *Injected {
	injected := &Injected{
		Insulin: datum.ToString(insulinField.Name, errs),
		Value:   datum.ToInt(valueField.Name, errs),
		Base:    b,
	}
	types.GetPlatformValidator().SetErrorReasons(failureReasons).Struct(injected, errs)
	return injected
}

func InsulinValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	insulin, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = insulinField.Allowed[insulin]
	return ok
}

func ValueValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	value, ok := field.Interface().(int)
	if !ok {
		return false
	}
	return value > valueField.LowerLimit
}
