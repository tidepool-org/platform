package data

import (
	"fmt"
	"reflect"
	"time"

	valid "github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
	"github.com/tidepool-org/platform/Godeps/_workspace/src/labix.org/v2/mgo/bson"

	"github.com/tidepool-org/platform/validate"
)

//used for all data types
var validator = validate.NewPlatformValidator()

func init() {
	validator.RegisterValidation(timeStrTag, TimeStringValidator)
	validator.RegisterValidation(timeObjTag, TimeObjectValidator)
	validator.RegisterValidation(timeZoneOffsetTag, TimezoneOffsetValidator)
	validator.RegisterValidation(payloadTag, PayloadValidator)
	validator.RegisterValidation(annotationsTag, AnnotationsValidator)
}

//Base represents tha base types that all device data records contain
type Base struct {
	//required data
	_ID      bson.ObjectId `bson:"_id" valid:"mongo,required"`
	ID       string        `json:"id" bson:"id" valid:"required"`
	UserID   string        `json:"userId" bson:"userId" valid:"required"`
	DeviceID string        `json:"deviceId" bson:"deviceId" valid:"required"`
	Time     string        `json:"time" bson:"time" valid:"timestr"`
	Type     string        `json:"type" bson:"type" valid:"required"`
	UploadID string        `json:"uploadId" bson:"uploadId" valid:"-"`

	//optional data
	DeviceTime       string        `json:"deviceTime,omitempty" bson:"deviceTime,omitempty" valid:"omitempty,timestr"`
	TimezoneOffset   int           `json:"timezoneOffset,omitempty" bson:"timezoneOffset,omitempty" valid:"omitempty,timezoneoffset"`
	ConversionOffset int           `json:"conversionOffset,omitempty" bson:"conversionOffset,omitempty" valid:"omitempty,required"`
	ClockDriftOffset int           `json:"clockDriftOffset,omitempty" bson:"clockDriftOffset,omitempty" valid:"omitempty,required"`
	Payload          interface{}   `json:"payload,omitempty" bson:"payload,omitempty" valid:"omitempty,payload"`
	Annotations      []interface{} `json:"annotations,omitempty" bson:"annotations,omitempty" valid:"omitempty,annotations"`

	//used for versioning and de-deping
	Internal `bson:",inline"`
}

//Internal are existing fields used for versioning and de-deping
type Internal struct {
	CreatedTime   string `json:"createdTime" bson:"createdTime" valid:"timestr"`
	GroupID       string `json:"-" bson:"_groupId" valid:"required"`
	ActiveFlag    bool   `json:"-" bson:"_active" valid:"required"`
	SchemaVersion int    `json:"-" bson:"_schemaVersion" valid:"required,min=0"`
	Version       int    `json:"-" bson:"_version,omitempty" valid:"-"`
}

var (
	//InternalFields are what we only use internally in the service and don't wish to return
	InternalFields = map[string]interface{}{
		"_groupId":       0,
		"_active":        0,
		"_schemaVersion": 0,
		"_version":       0,
		"createdTime":    0,
	}
)

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

	//validation
	tzValidationLowerLimit = -840
	tzValidationUpperLimit = 720
	timeStrValidationMsg   = "Times need to be ISO 8601 format and not in the future"

	timeStrTag        validate.ValidationTag = "timestr"
	timeObjTag        validate.ValidationTag = "timeobj"
	timeZoneOffsetTag validate.ValidationTag = "timezoneoffset"
	payloadTag        validate.ValidationTag = "payload"
	annotationsTag    validate.ValidationTag = "annotations"
)

//BuildBase builds the base fields that all device data records contain
func BuildBase(obj map[string]interface{}) (Base, *Error) {

	errs := NewError(obj)
	cast := NewCaster(errs)

	base := Base{
		_ID:      bson.NewObjectId(),
		ID:       bson.NewObjectId().Hex(),
		UserID:   cast.ToString(UserIDField, obj[UserIDField]),
		DeviceID: cast.ToString(deviceIDField, obj[deviceIDField]),
		UploadID: cast.ToString(uploadIDField, obj[uploadIDField]),
		Time:     cast.ToString(timeField, obj[timeField]),
		Type:     cast.ToString(typeField, obj[typeField]),
		Payload:  obj[payloadField],
		Internal: Internal{
			GroupID:       cast.ToString(GroupIDField, obj[GroupIDField]),
			ActiveFlag:    true,
			SchemaVersion: 1, //TODO: configured ??
			CreatedTime:   time.Now().Format(time.RFC3339),
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
	if obj[annotationsField] != nil {
		base.Annotations = cast.ToArray(annotationsField, obj[annotationsField])
	}

	if validationErrors := validator.Struct(base); len(validationErrors) > 0 {
		errs.AppendError(validationErrors.GetError(validationFailureReasons))
	}

	return base, errs
}

var validationFailureReasons = validate.ErrorReasons{
	timeStrTag:        timeStrValidationMsg,
	timeZoneOffsetTag: fmt.Sprintf("TimezoneOffset needs to be in minutes and greater than %d and less than %d", tzValidationLowerLimit, tzValidationUpperLimit),
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
	if _, ok := field.Interface().([]interface{}); ok {
		return true
	}
	return false
}

func TimezoneOffsetValidator(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	if offset, ok := field.Interface().(int); ok {
		if offset >= tzValidationLowerLimit && offset <= tzValidationUpperLimit {
			return true
		}
	}
	return false
}

func TimeObjectValidator(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	if timeObject, ok := field.Interface().(time.Time); ok {
		return isTimeObjectValid(timeObject)
	}
	return false
}

func isTimeObjectValid(timeObject time.Time) bool {
	return !timeObject.IsZero() && timeObject.Before(time.Now())
}

func TimeStringValidator(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	if timeString, ok := field.Interface().(string); ok {
		return isTimeStringValid(timeString)
	}
	return false
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

//ToArray will return the given data as []interface{}
func (cast *Cast) ToArray(fieldName string, data interface{}) []interface{} {
	if data == nil {
		return nil
	}
	arrayData, ok := data.([]interface{})
	if !ok {
		cast.err.AppendFieldError(fieldName, data)
		return nil
	}
	return arrayData
}
