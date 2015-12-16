package data

import (
	"github.com/tidepool-org/platform/validate"
)

type Base struct {
	Type             string  `json:"type" valid:"required"`
	DeviceTime       string  `json:"deviceTime" valid:"required"`
	Time             string  `json:"time" valid:"required"`
	TimezoneOffset   float64 `json:"timezoneOffset" valid:"required"`
	ConversionOffset float64 `json:"conversionOffset" valid:"required"`
	DeviceId         string  `json:"deviceId" valid:"required"`
}

var validator = validate.PlatformValidator{}

func buildBase(obj map[string]interface{}) (Base, error) {
	const (
		type_field              = "type"
		device_time_field       = "deviceTime"
		timezone_offset_field   = "timezoneOffset"
		time_field              = "time"
		conversion_offset_field = "conversionOffset"
		device_id_field         = "deviceId"
	)

	base := Base{
		ConversionOffset: obj[conversion_offset_field].(float64),
		TimezoneOffset:   obj[timezone_offset_field].(float64),
		DeviceId:         obj[device_id_field].(string),
		DeviceTime:       obj[device_time_field].(string),
		Time:             obj[time_field].(string),
		Type:             obj[type_field].(string),
	}

	_, err := validator.Validate(base)
	return base, err
}

func GetData() string {
	return "data"
}
