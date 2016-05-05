package types

import (
	"reflect"
	"time"

	validator "gopkg.in/bluesuncorp/validator.v8"
)

func init() {
	GetPlatformValidator().RegisterValidation(ZuluTimeStringField.Tag, ZuluPastTimeStringValidator)
	GetPlatformValidator().RegisterValidation(NonZuluTimeStringField.Tag, NonZuluPastTimeStringValidator)
	GetPlatformValidator().RegisterValidation(OffsetTimeStringField.Tag, OffsetPastTimeStringValidator)
	GetPlatformValidator().RegisterValidation(OffsetOrZuluTimeStringField.Tag, OffsetOrZuluPastTimeStringValidator)
}

var (
	ZuluTimeStringField = DateDatumField{
		DatumField: &DatumField{Name: "time"},
		Tag:        "zuluTimeString",
		Message:    "An ISO 8601-formatted UTC timestamp with a final Z for 'Zulu' time e.g 2013-05-04T03:58:44.584Z",
		AllowedDate: &AllowedDate{
			//refer to https://golang.org/pkg/time/#pkg-constants
			Format:        "2006-01-02T15:04:05Z",
			LowerLimit:    "2007-01-01T00:00:00Z",
			AllowedFuture: false, // TODO_DATA: Do not allow future dates
		},
	}

	NonZuluTimeStringField = DateDatumField{
		DatumField: &DatumField{Name: "time"},
		Tag:        "nonZuluTimeString",
		Message:    "An ISO 8601 formatted timestamp without any timezone offset information e.g 2013-05-04T03:58:44.584",
		AllowedDate: &AllowedDate{
			//refer to https://golang.org/pkg/time/#pkg-constants
			Format:        "2006-01-02T15:04:05",
			LowerLimit:    "2007-01-01T00:00:00",
			AllowedFuture: true, // TODO_DATA: Allow future dates
		},
	}

	OffsetTimeStringField = DateDatumField{
		DatumField: &DatumField{Name: "time"},
		Tag:        "offsetTimeString",
		Message:    "An ISO 8601-formatted timestamp including a timezone offset from UTC e.g 2013-05-04T03:58:44-08:00",
		AllowedDate: &AllowedDate{
			//refer to https://golang.org/pkg/time/#pkg-constants
			Format:        "2006-01-02T15:04:05-07:00",
			LowerLimit:    "2007-01-01T00:00:00-00:00",
			AllowedFuture: false, // TODO_DATA: Do not allow future dates
		},
	}

	OffsetOrZuluTimeStringField = DateDatumField{
		DatumField: &DatumField{Name: "time"},
		Tag:        "offsetOrZuluTimeString",
		Message:    "An ISO 8601-formatted timestamp including either a timezone offset from UTC OR converted to UTC with a final Z for 'Zulu' time. e.g.2013-05-04T03:58:44.584Z OR 2013-05-04T03:58:44-08:00",
	}
)

func NonZuluPastTimeStringValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {

	timeString, ok := field.Interface().(string)
	if !ok {
		return false
	}
	return isTimeStringValid(timeString, NonZuluTimeStringField)
}

func ZuluPastTimeStringValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {

	timeString, ok := field.Interface().(string)
	if !ok {
		return false
	}
	return isTimeStringValid(timeString, ZuluTimeStringField)
}

func OffsetOrZuluPastTimeStringValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {

	timeString, ok := field.Interface().(string)
	if !ok {
		return false
	}
	if isTimeStringValid(timeString, OffsetTimeStringField) {
		return true
	}
	return isTimeStringValid(timeString, ZuluTimeStringField)
}

func OffsetPastTimeStringValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {

	timeString, ok := field.Interface().(string)
	if !ok {
		return false
	}
	return isTimeStringValid(timeString, OffsetTimeStringField)
}

func isTimeStringValid(timeString string, dateDatum DateDatumField) bool {
	var timeObject time.Time
	timeObject, err := time.Parse(dateDatum.Format, timeString)
	if err != nil {
		return false
	}

	// TODO_DATA: Are future dates allowed?
	if !dateDatum.AllowedFuture && timeObject.After(time.Now()) {
		return false
	}

	lowerTimeObject, err := time.Parse(dateDatum.Format, dateDatum.LowerLimit)
	if err != nil {
		return false
	}

	return lowerTimeObject.Before(timeObject)

}
