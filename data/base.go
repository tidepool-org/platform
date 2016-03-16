package data

import (
	"reflect"
	"time"

	valid "github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
	"github.com/tidepool-org/platform/Godeps/_workspace/src/labix.org/v2/mgo/bson"

	"github.com/tidepool-org/platform/validate"
)

//used for all data types
var validator = validate.NewPlatformValidator()

func init() {
	validator.RegisterValidation("timestr", TimeStringValidator)
	validator.RegisterValidation("timeobj", TimeObjectValidator)
	validator.RegisterValidation("timezoneoffset", TimezoneOffsetValidator)
	validator.RegisterValidation("payload", PayloadValidator)
	validator.RegisterValidation("annotations", AnnotationsValidator)
}

//Base represents tha base types that all device data records contain
type Base struct {
	//required data
	_ID         bson.ObjectId `bson:"_id" valid:"mongo,required"`
	ID          string        `json:"id" bson:"id" valid:"required"`
	UserID      string        `json:"userId" bson:"userId" valid:"required"`
	DeviceID    string        `json:"deviceId" bson:"deviceId" valid:"required"`
	Time        string        `json:"time" bson:"time" valid:"timestr"`
	Type        string        `json:"type" bson:"type" valid:"required"`
	UploadID    string        `json:"uploadId" bson:"uploadId" valid:"-"`
	CreatedTime string        `json:"createdTime" bson:"createdTime" valid:"timestr"`

	//optional data
	DeviceTime       string      `json:"deviceTime,omitempty" bson:"deviceTime,omitempty" valid:"omitempty,timestr"`
	TimezoneOffset   int         `json:"timezoneOffset,omitempty" bson:"timezoneOffset,omitempty" valid:"omitempty,timezoneoffset"`
	ConversionOffset int         `json:"conversionOffset,omitempty" bson:"conversionOffset,omitempty" valid:"omitempty,required"`
	ClockDriftOffset int         `json:"clockDriftOffset,omitempty" bson:"clockDriftOffset,omitempty" valid:"omitempty,required"`
	Payload          interface{} `json:"payload,omitempty" bson:"payload,omitempty" valid:"omitempty,payload"`
	Annotations      interface{} `json:"annotations,omitempty" bson:"annotations,omitempty" valid:"omitempty,annotations"`

	//used for versioning and de-deping
	Storage `bson:",inline"`
}

//Storage are existing fields used for versioning and de-deping
type Storage struct {
	GroupID       string `json:"-" bson:"_groupId" valid:"required"`
	ActiveFlag    bool   `json:"-" bson:"_active" valid:"required"`
	SchemaVersion int    `json:"-" bson:"_schemaVersion" valid:"required,min=0"`
	Version       int    `json:"-" bson:"_version,omitempty" valid:"-"`
}

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

//BuildBase builds the base fields that all device data records contain
func BuildBase(obj map[string]interface{}) (Base, *Error) {

	errs := NewError(obj)
	cast := NewCaster(errs)

	base := Base{
		_ID:         bson.NewObjectId(),
		ID:          bson.NewObjectId().Hex(),
		CreatedTime: time.Now().Format(time.RFC3339),
		UserID:      cast.ToString(UserIDField, obj[UserIDField]),
		DeviceID:    cast.ToString(deviceIDField, obj[deviceIDField]),
		UploadID:    cast.ToString(uploadIDField, obj[uploadIDField]),
		Time:        cast.ToString(timeField, obj[timeField]),
		Type:        cast.ToString(typeField, obj[typeField]),
		Payload:     obj[payloadField],
		Annotations: obj[annotationsField],
		Storage: Storage{
			GroupID:       cast.ToString(GroupIDField, obj[GroupIDField]),
			ActiveFlag:    true,
			SchemaVersion: 1, //TODO: configured ??
		},
	}

	//set optional data
	if obj[conversionOffsetField] != nil {
		base.ConversionOffset = cast.ToInt(conversionOffsetField, obj[conversionOffsetField])
	}
	if obj[conversionOffsetField] != nil {
		base.ConversionOffset = cast.ToInt(conversionOffsetField, obj[conversionOffsetField])
	}
	if obj[timezoneOffsetField] != nil {
		base.TimezoneOffset = cast.ToInt(timezoneOffsetField, obj[timezoneOffsetField])
	}
	if obj[timezoneOffsetField] != nil {
		base.ClockDriftOffset = cast.ToInt(clockDriftOffsetField, obj[clockDriftOffsetField])
	}
	if obj[timezoneOffsetField] != nil {
		base.DeviceTime = cast.ToString(deviceTimeField, obj[deviceTimeField])
	}

	errs.AppendError(validator.ValidateStruct(base))
	return base, errs
}

func PayloadValidator(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	//a place holder for more through validation
	if _, ok := field.Interface().(interface{}); ok {
		return true
	}
	return false
}

func AnnotationsValidator(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	//a place holder for more through validation
	if _, ok := field.Interface().(interface{}); ok {
		return true
	}
	return false
}

func TimezoneOffsetValidator(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	if offset, ok := field.Interface().(int); ok {
		if offset >= -840 && offset <= 720 {
			return true
		}
	}
	return false
}

func TimeObjectValidator(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	if timeObject, ok := field.Interface().(time.Time); !ok {
		return false
	} else {
		return isTimeObjectValid(timeObject)
	}
}

func isTimeObjectValid(timeObject time.Time) bool {
	return !timeObject.IsZero() && timeObject.Before(time.Now())
}

func TimeStringValidator(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	if timeString, ok := field.Interface().(string); !ok {
		return false
	} else {
		return isTimeStringValid(timeString)
	}
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
