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
	cast := NewCaster(errs)

	base := Base{
		ConversionOffset: cast.ToInt(conversion_offset_field, obj[conversion_offset_field]),
		TimezoneOffset:   cast.ToInt(timezone_offset_field, obj[timezone_offset_field]),
		DeviceId:         cast.ToString(device_id_field, obj[device_id_field]),
		DeviceTime:       cast.ToTime(device_time_field, obj[device_time_field]),
		Time:             cast.ToTime(time_field, obj[time_field]),
		Type:             cast.ToString(type_field, obj[type_field]),
	}

	_, err := validator.Validate(base)
	errs.AppendError(err)
	return base, errs
}

func GetData() string {
	return "data"
}

type Cast struct {
	err *DataError
}

func NewCaster(err *DataError) *Cast {
	return &Cast{err: err}
}

func (this *Cast) ToString(fieldName string, data interface{}) string {
	aString, ok := data.(string)
	if !ok {
		this.err.AppendFieldError(fieldName, data)
	}
	return aString
}

func (this *Cast) ToFloat64(fieldName string, data interface{}) float64 {
	theFloat, ok := data.(float64)
	if !ok {
		this.err.AppendFieldError(fieldName, data)
	}
	return theFloat
}

func (this *Cast) ToInt(fieldName string, data interface{}) int {
	theInt, ok := data.(int)
	if !ok {
		theFloat := this.ToFloat64(fieldName, data)
		theInt = int(theFloat)
	}
	return theInt
}

func (this *Cast) ToTime(fieldName string, data interface{}) time.Time {
	timeStr, ok := data.(string)
	if ok {
		theTime, err := time.Parse(time.RFC3339, timeStr)
		this.err.AppendError(err)
		return theTime
	}
	this.err.AppendFieldError(fieldName, data)
	return time.Time{}
}
