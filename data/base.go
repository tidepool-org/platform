package data

import (
	"time"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/labix.org/v2/mgo/bson"

	"github.com/tidepool-org/platform/validate"
)

//Base represents tha base types that all device data records contain
type Base struct {
	ID               bson.ObjectId `json:"_id" bson:"_id"`
	UserID           string        `json:"userId" bson:"userId" valid:"required"`
	Type             string        `json:"type" bson:"type" valid:"required"`
	DeviceTime       time.Time     `json:"deviceTime" bson:"deviceTime" valid:"required"`
	Time             time.Time     `json:"time" bson:"time" valid:"required"`
	TimezoneOffset   int           `json:"timezoneOffset" bson:"timezoneOffset,omitempty"`
	ConversionOffset int           `json:"conversionOffset" bson:"conversionOffset,omitempty"`
	DeviceID         string        `json:"deviceId" bson:"deviceId" valid:"required"`
}

var validator = validate.PlatformValidator{}

//BuildBase builds the base fields that all device data records contain
func BuildBase(obj map[string]interface{}) (Base, *Error) {
	const (
		useridField           = "userId"
		typeField             = "type"
		deviceTimeField       = "deviceTime"
		timezoneOffsetField   = "timezoneOffset"
		timeField             = "time"
		conversionOffsetField = "conversionOffset"
		deviceIDField         = "deviceId"
	)

	errs := NewError(obj)
	cast := NewCaster(errs)

	base := Base{
		ID:               bson.NewObjectId(),
		UserID:           cast.ToString(useridField, obj[useridField]),
		ConversionOffset: cast.ToInt(conversionOffsetField, obj[conversionOffsetField]),
		TimezoneOffset:   cast.ToInt(timezoneOffsetField, obj[timezoneOffsetField]),
		DeviceID:         cast.ToString(deviceIDField, obj[deviceIDField]),
		DeviceTime:       cast.ToTime(deviceTimeField, obj[deviceTimeField]),
		Time:             cast.ToTime(timeField, obj[timeField]),
		Type:             cast.ToString(typeField, obj[typeField]),
	}

	_, err := validator.ValidateStruct(base)
	errs.AppendError(err)
	return base, errs
}

//Cast type for use in casting our incoming generic json data to the expected types that our data model uses
type Cast struct {
	err *Error
}

//NewCaster creates a Cast
func NewCaster(err *Error) *Cast {
	return &Cast{err: err}
}

//ToString will return the given data as a string or add an error to the cast obj
func (cast *Cast) ToString(fieldName string, data interface{}) string {
	aString, ok := data.(string)
	if !ok {
		cast.err.AppendFieldError(fieldName, data)
	}
	return aString
}

//ToFloat64 will return the given data as a float64 or add an error to the cast obj
func (cast *Cast) ToFloat64(fieldName string, data interface{}) float64 {
	theFloat, ok := data.(float64)
	if !ok {
		cast.err.AppendFieldError(fieldName, data)
	}
	return theFloat
}

//ToInt will return the given data as a int or add an error to the cast obj
func (cast *Cast) ToInt(fieldName string, data interface{}) int {
	theInt, ok := data.(int)
	if !ok {
		theFloat := cast.ToFloat64(fieldName, data)
		theInt = int(theFloat)
	}
	return theInt
}

//ToTime will return the given data as time.Time or add an error to the cast obj
func (cast *Cast) ToTime(fieldName string, data interface{}) time.Time {
	timeStr, ok := data.(string)
	if ok {
		theTime, err := time.Parse(time.RFC3339, timeStr)
		cast.err.AppendError(err)
		return theTime
	}
	cast.err.AppendFieldError(fieldName, data)
	return time.Time{}
}
