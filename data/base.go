package data

import (
	"time"

	"github.com/tidepool-org/platform/validate"
)

type Base struct {
	Type             string    `json:"type" valid:"required"`
	DeviceTime       time.Time `json:"deviceTime" valid:"required"`
	Time             time.Time `json:"time" valid:"required"`
	TimezoneOffset   int       `json:"timezoneOffset"`
	ConversionOffset int       `json:"conversionOffset"`
	DeviceId         string    `json:"deviceId" valid:"required"`
}

var validator = validate.PlatformValidator{}

func getTime(name string, detail interface{}, e *DataError) time.Time {
	timeStr, ok := detail.(string)
	if ok {
		theTime, err := time.Parse(time.RFC3339, timeStr)
		e.AppendError(err)
		return theTime
	}
	e.AppendFieldError(name, detail)
	return time.Time{}
}

func buildBase(obj map[string]interface{}) (Base, *DataError) {
	const (
		type_field              = "type"
		device_time_field       = "deviceTime"
		timezone_offset_field   = "timezoneOffset"
		time_field              = "time"
		conversion_offset_field = "conversionOffset"
		device_id_field         = "deviceId"
	)

	errs := NewDataError(obj)

	conversionOffset, ok := obj[conversion_offset_field].(int)
	if !ok {
		conversionOffset_float64, ok := obj[conversion_offset_field].(float64)
		conversionOffset = int(conversionOffset_float64)
		if !ok {
			errs.AppendFieldError(conversion_offset_field, obj[conversion_offset_field])
		}
	}

	timezoneOffset, ok := obj[timezone_offset_field].(int)
	if !ok {
		timezoneOffset_float64, ok := obj[timezone_offset_field].(float64)
		timezoneOffset = int(timezoneOffset_float64)
		if !ok {
			errs.AppendFieldError(timezone_offset_field, obj[timezone_offset_field])
		}
	}

	deviceId, ok := obj[device_id_field].(string)
	if !ok {
		errs.AppendFieldError(device_id_field, obj[device_id_field])
	}

	deviceTime := getTime(device_time_field, obj[device_time_field], errs)
	eventTime := getTime(time_field, obj[time_field], errs)

	typeOf, ok := obj[type_field].(string)
	if !ok {
		errs.AppendFieldError(type_field, obj[type_field])
	}

	base := Base{
		ConversionOffset: conversionOffset,
		TimezoneOffset:   timezoneOffset,
		DeviceId:         deviceId,
		DeviceTime:       deviceTime,
		Time:             eventTime,
		Type:             typeOf,
	}

	_, err := validator.Validate(base)
	errs.AppendError(err)
	return base, errs
}

func GetData() string {
	return "data"
}
