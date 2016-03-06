package data

import (
	"time"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/labix.org/v2/mgo/bson"

	"github.com/tidepool-org/platform/validate"
)

//Base represents tha base types that all device data records contain
type Base struct {
	ID               string      `json:"id" bson:"_id" valid:"required"`
	UserID           string      `json:"userId" bson:"userId" valid:"required"`
	DeviceID         string      `json:"deviceId" bson:"deviceId" valid:"required"`
	UploadID         string      `json:"uploadId" bson:"uploadId" valid:"-"`
	DeviceTime       string      `json:"deviceTime" bson:"deviceTime" valid:"required"`
	Time             string      `json:"time" bson:"time" valid:"required"`
	TimezoneOffset   int         `json:"timezoneOffset" bson:"timezoneOffset,omitempty" valid:"-"`
	ConversionOffset int         `json:"conversionOffset" bson:"conversionOffset,omitempty" valid:"-"`
	ClockDriftOffset int         `json:"clockDriftOffset" bson:"clockDriftOffset,omitempty" valid:"-"`
	Type             string      `json:"type" bson:"type" valid:"required"`
	Payload          interface{} `json:"payload" bson:"payload,omitempty" valid:"-"`
	Annotations      interface{} `json:"annotations" bson:"annotations,omitempty" valid:"-"`
	BaseDataStorage  `bson:",inline"`
}

//BaseDataStorage are existing fields used for verioning and de-deping
type BaseDataStorage struct {
	GroupID       string `json:"-" bson:"_groupId" valid:"required"`
	ActiveFlag    bool   `json:"-" bson:"_active" valid:"required"`
	SchemaVersion int    `json:"-" bson:"_schemaVersion" valid:"required"`
	Version       int    `json:"-" bson:"_version,omitempty" valid:"-"`
	CreatedTime   string `json:"createdTime" bson:"createdTime" valid:"required"`
}

var validator = validate.PlatformValidator{}

const (
	userIDField   = "userId"
	deviceIDField = "deviceId"
	uploadIDField = "uploadId"

	timezoneOffsetField   = "timezoneOffset"
	conversionOffsetField = "conversionOffset"
	clockDriftOffsetField = "clockDriftOffset"

	typeField       = "type"
	timeField       = "time"
	deviceTimeField = "deviceTime"

	payloadField     = "payload"
	annotationsField = "annotations"
)

//BuildBase builds the base fields that all device data records contain
func BuildBase(obj map[string]interface{}) (Base, *Error) {

	errs := NewError(obj)
	cast := NewCaster(errs)

	base := Base{
		ID:               bson.NewObjectId().Hex(),
		UserID:           cast.ToString(userIDField, obj[userIDField]),
		DeviceID:         cast.ToString(deviceIDField, obj[deviceIDField]),
		UploadID:         cast.ToString(uploadIDField, obj[uploadIDField]),
		ConversionOffset: cast.ToInt(conversionOffsetField, obj[conversionOffsetField]),
		TimezoneOffset:   cast.ToInt(timezoneOffsetField, obj[timezoneOffsetField]),
		ClockDriftOffset: cast.ToInt(clockDriftOffsetField, obj[clockDriftOffsetField]),
		DeviceTime:       cast.ToString(deviceTimeField, obj[deviceTimeField]),
		Time:             cast.ToString(timeField, obj[timeField]),
		Type:             cast.ToString(typeField, obj[typeField]),
		Payload:          obj[payloadField],
		Annotations:      obj[annotationsField],
		BaseDataStorage: BaseDataStorage{
			GroupID:       "85e9e57e20", //cast.ToString(useridField, obj[useridField]),
			ActiveFlag:    true,
			SchemaVersion: 1,
			CreatedTime:   time.Now().Format(time.RFC3339),
		},
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
	if data == nil {
		return ""
	}
	aString, ok := data.(string)
	if !ok {
		cast.err.AppendFieldError(fieldName, data)
	}
	return aString
}

//ToFloat64 will return the given data as a float64 or add an error to the cast obj
func (cast *Cast) ToFloat64(fieldName string, data interface{}) float64 {
	if data == nil {
		return 0.0
	}
	theFloat, ok := data.(float64)
	if !ok {
		cast.err.AppendFieldError(fieldName, data)
	}
	return theFloat
}

//ToInt will return the given data as a int or add an error to the cast obj
func (cast *Cast) ToInt(fieldName string, data interface{}) int {
	if data == nil {
		return 0
	}
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
		if err != nil {
			//try this format also before we fail
			theTime, err = time.Parse("2006-01-02T15:04:05", timeStr)
		}

		cast.err.AppendError(err)
		return theTime
	}
	cast.err.AppendFieldError(fieldName, data)
	return time.Time{}
}
