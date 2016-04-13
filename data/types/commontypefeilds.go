package types

import (
	"fmt"
	"reflect"
	"time"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
)

func init() {
	GetPlatformValidator().RegisterValidation(BloodGlucoseValueField.Tag, BloodGlucoseValueValidator)
	GetPlatformValidator().RegisterValidation(MmolOrMgUnitsField.Tag, MmolOrMgUnitsValidator)
	GetPlatformValidator().RegisterValidation(MmolUnitsField.Tag, MmolUnitsValidator)
	GetPlatformValidator().RegisterValidation(BolusSubTypeField.Tag, BolusSubTypeValidator)
	GetPlatformValidator().RegisterValidation(TimeStringField.Tag, PastTimeStringValidator)
	GetPlatformValidator().RegisterValidation("timeobject", PastTimeObjectValidator)
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

	BloodGlucoseValueField = FloatDatumField{
		DatumField:        &DatumField{Name: "value"},
		Tag:               "bloodglucosevalue",
		Message:           "Must be greater than 0.0",
		AllowedFloatRange: &AllowedFloatRange{LowerLimit: 0.0},
	}

	BolusSubTypeField = DatumFieldInformation{
		DatumField: &DatumField{Name: "subType"},
		Tag:        "bolussubtype",
		Message:    "Must be one of normal, square, dual/square",
		Allowed:    Allowed{"normal": true, "square": true, "dual/square": true},
	}

	TimeStringField = DatumFieldInformation{
		DatumField: &DatumField{Name: "time"},
		Tag:        "timestr",
		Message:    "Times need to be ISO 8601 format and not in the future",
	}
)

func BolusSubTypeValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	subType, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = BolusSubTypeField.Allowed[subType]
	return ok
}

func BloodGlucoseValueValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	val, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return val > BloodGlucoseValueField.LowerLimit
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

func PastTimeObjectValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	timeObject, ok := field.Interface().(time.Time)
	if !ok {
		return false
	}
	return isTimeObjectValid(timeObject)
}

func isTimeObjectValid(timeObject time.Time) bool {
	return !timeObject.IsZero() && timeObject.Before(time.Now())
}

func PastTimeStringValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {

	timeString, ok := field.Interface().(string)
	if !ok {
		return false
	}
	valid := isTimeStringValid(timeString)
	return valid
}

func isTimeStringValid(timeString string) bool {
	var timeObject time.Time
	timeObject, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		timeObject, err = time.Parse("2006-01-02T15:04:05", timeString)
		if err != nil {
			return false
		}
	}

	return isTimeObjectValid(timeObject)
}
