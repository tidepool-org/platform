package types

import (
	"fmt"
	"reflect"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
)

func init() {
	GetPlatformValidator().RegisterValidation(BloodValueField.Tag, BloodValueValidator)
	GetPlatformValidator().RegisterValidation(MmolOrMgUnitsField.Tag, MmolOrMgUnitsValidator)
	GetPlatformValidator().RegisterValidation(MmolUnitsField.Tag, MmolUnitsValidator)
}

var (
	mmol = "mmol/L"
	mg   = "mg/dL"

	MmolOrMgUnitsField = DatumFieldInformation{
		DatumField: &DatumField{Name: "units"},
		Tag:        "mmolmgunits",
		Message:    fmt.Sprintf("Must be one of %s, %s", mmol, mg),
		Allowed: Allowed{
			mmol:     true,
			"mmol/l": true,
			mg:       true,
			"mg/dl":  true,
		},
	}

	MmolUnitsField = DatumFieldInformation{
		DatumField: &DatumField{Name: "units"},
		Tag:        "mmolunits",
		Message:    fmt.Sprintf("Must be %s", mmol),
		Allowed: Allowed{
			mmol:     true,
			"mmol/l": true,
		},
	}

	BloodValueField = FloatDatumField{
		DatumField:        &DatumField{Name: "value"},
		Tag:               "bloodvalue",
		Message:           "Must be greater than 0.0",
		AllowedFloatRange: &AllowedFloatRange{LowerLimit: 0.0},
	}
)

func BloodValueValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	val, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return val > BloodValueField.LowerLimit
}

func MmolUnitsValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	units, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = MmolUnitsField.Allowed[units]
	return ok
}

func MmolOrMgUnitsValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	units, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = MmolOrMgUnitsField.Allowed[units]
	return ok
}
