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

func buildBase(t map[string]interface{}) (Base, error) {
	const (
		typeField             = "type"
		deviceTimeField       = "deviceTime"
		timezoneOffsetField   = "timezoneOffset"
		timeField             = "time"
		conversionOffsetField = "conversionOffset"
		deviceIdField         = "deviceId"
	)

	base := Base{
		ConversionOffset: t[conversionOffsetField].(float64),
		TimezoneOffset:   t[timezoneOffsetField].(float64),
		DeviceId:         t[deviceIdField].(string),
		DeviceTime:       t[deviceTimeField].(string),
		Time:             t[timeField].(string),
		Type:             t[typeField].(string),
	}

	_, err := validator.Validate(base)
	return base, err
}

func GetData() string {
	return "data"
}
