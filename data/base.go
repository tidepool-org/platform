package data

import (
	"time"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/labix.org/v2/mgo/bson"

	"github.com/tidepool-org/platform/validate"
)

//Base represents tha base types that all device data records contain
type Base struct {
	//required data
	_ID              bson.ObjectId `bson:"_id" valid:"mongo,required"`
	ID               string        `json:"id" bson:"id" valid:"required"`
	UserID           string        `json:"userId" bson:"userId" valid:"required"`
	DeviceID         string        `json:"deviceId" bson:"deviceId" valid:"required"`
	Time             string        `json:"time" bson:"time" valid:"required"`
	Type             string        `json:"type" bson:"type" valid:"required"`
	UploadID         string        `json:"uploadId" bson:"uploadId" valid:"-"`
	CreatedTime      string        `json:"createdTime" bson:"createdTime" valid:"required"`
	OptionalBaseData `bson:",inline"`
	BaseDataStorage  `bson:",inline"`
}

//OptionalBaseData are fields that if they exist we save them otherwise they are omitted
type OptionalBaseData struct {
	DeviceTime       string      `json:"deviceTime,omitempty" bson:"deviceTime,omitempty" valid:"-"`
	TimezoneOffset   int         `json:"timezoneOffset,omitempty" bson:"timezoneOffset,omitempty" valid:"-"`
	ConversionOffset int         `json:"conversionOffset,omitempty" bson:"conversionOffset,omitempty" valid:"-"`
	ClockDriftOffset int         `json:"clockDriftOffset,omitempty" bson:"clockDriftOffset,omitempty" valid:"-"`
	Payload          interface{} `json:"payload,omitempty" bson:"payload,omitempty" valid:"-"`
	Annotations      interface{} `json:"annotations,omitempty" bson:"annotations,omitempty" valid:"-"`
}

//BaseDataStorage are existing fields used for versioning and de-deping
type BaseDataStorage struct {
	GroupID       string `json:"-" bson:"_groupId" valid:"required"`
	ActiveFlag    bool   `json:"-" bson:"_active" valid:"required"`
	SchemaVersion int    `json:"-" bson:"_schemaVersion" valid:"required"`

	Version int `json:"-" bson:"_version,omitempty" valid:"-"`
}

var validator = validate.PlatformValidator{}

const (
	//UserIDField is the userID
	UserIDField = "userId"
	//GroupIDField id the groupID
	GroupIDField = "groupId"

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

func initOptionalBaseData(obj map[string]interface{}, caster *Cast) OptionalBaseData {
	optional := OptionalBaseData{}

	if obj[conversionOffsetField] != nil {
		optional.ConversionOffset = caster.ToInt(conversionOffsetField, obj[conversionOffsetField])
	}
	if obj[conversionOffsetField] != nil {
		optional.ConversionOffset = caster.ToInt(conversionOffsetField, obj[conversionOffsetField])
	}
	if obj[timezoneOffsetField] != nil {
		optional.TimezoneOffset = caster.ToInt(timezoneOffsetField, obj[timezoneOffsetField])
	}
	if obj[timezoneOffsetField] != nil {
		optional.ClockDriftOffset = caster.ToInt(clockDriftOffsetField, obj[clockDriftOffsetField])
	}
	if obj[timezoneOffsetField] != nil {
		optional.DeviceTime = caster.ToString(deviceTimeField, obj[deviceTimeField])
	}
	optional.Payload = obj[payloadField]
	optional.Annotations = obj[annotationsField]
	return optional
}

//BuildBase builds the base fields that all device data records contain
func BuildBase(obj map[string]interface{}) (Base, *Error) {

	errs := NewError(obj)
	cast := NewCaster(errs)

	base := Base{
		_ID:              bson.NewObjectId(),
		ID:               bson.NewObjectId().Hex(),
		CreatedTime:      time.Now().Format(time.RFC3339),
		UserID:           cast.ToString(UserIDField, obj[UserIDField]),
		DeviceID:         cast.ToString(deviceIDField, obj[deviceIDField]),
		UploadID:         cast.ToString(uploadIDField, obj[uploadIDField]),
		Time:             cast.ToString(timeField, obj[timeField]),
		Type:             cast.ToString(typeField, obj[typeField]),
		OptionalBaseData: initOptionalBaseData(obj, cast),
		BaseDataStorage: BaseDataStorage{
			GroupID:       cast.ToString(GroupIDField, obj[GroupIDField]),
			ActiveFlag:    true,
			SchemaVersion: 1,
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
