package types

import (
	"reflect"
	"time"

	validator "gopkg.in/bluesuncorp/validator.v8"
)

func init() {
	GetPlatformValidator().RegisterValidation(TimeStringField.Tag, PastTimeStringValidator)
	GetPlatformValidator().RegisterValidation("timeobject", PastTimeObjectValidator)
}

var TimeStringField = DatumFieldInformation{
	DatumField: &DatumField{Name: "time"},
	Tag:        "timestr",
	Message:    "Times need to be ISO 8601 format and not in the future",
}

//2007-01-01T00:00:00

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
