package data

import "github.com/tidepool-org/platform/validate"

type Base struct {
	Type             string `json:"type" valid:"required"`
	DeviceTime       string `json:"deviceTime" valid:"required"`
	Time             string `json:"time" valid:"required"`
	TimezoneOffset   int    `json:"timezoneOffset" valid:"required"`
	ConversionOffset int    `json:"conversionOffset" valid:"required"`
	DeviceId         string `json:"deviceId" valid:"required"`
}

var validator = validate.PlatformValidator{}

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
		errs.AppendFieldError(conversion_offset_field, obj[conversion_offset_field])
	}

	timezoneOffset, ok := obj[timezone_offset_field].(int)
	if !ok {
		errs.AppendFieldError(timezone_offset_field, obj[timezone_offset_field])
	}

	deviceId, ok := obj[device_id_field].(string)
	if !ok {
		errs.AppendFieldError(device_id_field, obj[device_id_field])
	}

	deviceTime, ok := obj[device_time_field].(string)
	if !ok {
		errs.AppendFieldError(device_time_field, obj[device_time_field])
	}

	time, ok := obj[time_field].(string)
	if !ok {
		errs.AppendFieldError(time_field, obj[time_field])
	}

	typeOf, ok := obj[type_field].(string)
	if !ok {
		errs.AppendFieldError(type_field, obj[type_field])
	}

	base := Base{
		ConversionOffset: conversionOffset,
		TimezoneOffset:   timezoneOffset,
		DeviceId:         deviceId,
		DeviceTime:       deviceTime,
		Time:             time,
		Type:             typeOf,
	}

	_, err := validator.Validate(base)
	errs.AppendError(err)
	return base, errs
}

func GetData() string {
	return "data"
}
