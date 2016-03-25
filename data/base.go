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
var validator = validate.NewPlatformValidator(validationFailureReasons)

func init() {
	validator.RegisterValidation(timeStringTag, TimeStringValidator)
	validator.RegisterValidation(timeObjectTag, TimeObjectValidator)
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
	DeviceTime       *string       `json:"deviceTime,omitempty" bson:"deviceTime,omitempty" valid:"omitempty,timestr"`
	TimezoneOffset   *int          `json:"timezoneOffset,omitempty" bson:"timezoneOffset,omitempty" valid:"omitempty,timezoneoffset"`
	ConversionOffset *int          `json:"conversionOffset,omitempty" bson:"conversionOffset,omitempty" valid:"omitempty,required"`
	ClockDriftOffset *int          `json:"clockDriftOffset,omitempty" bson:"clockDriftOffset,omitempty" valid:"omitempty,required"`
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
	UserIDField  = "userId"
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

	tzValidationLowerLimit = -840
	tzValidationUpperLimit = 720
	timeStrValidationMsg   = "Times need to be ISO 8601 format and not in the future"

	timeStringTag     validate.ValidationTag = "timestr"
	timeObjectTag     validate.ValidationTag = "timeobject"
	timeZoneOffsetTag validate.ValidationTag = "timezoneoffset"
	payloadTag        validate.ValidationTag = "payload"
	annotationsTag    validate.ValidationTag = "annotations"
)

func BuildBase(datum Datum, errs *validate.ErrorsArray) Base {

	base := Base{
		_ID:      bson.NewObjectId(),
		ID:       bson.NewObjectId().Hex(),
		UserID:   datum[UserIDField].(string),
		DeviceID: datum[deviceIDField].(string),
		UploadID: datum[uploadIDField].(string),
		Time:     datum[timeField].(string),
		Type:     datum[typeField].(string),
		Payload:  datum[payloadField],
		Internal: Internal{
			GroupID:       datum[GroupIDField].(string),
			ActiveFlag:    true,
			SchemaVersion: 1, //TODO: configured ??
			CreatedTime:   time.Now().Format(time.RFC3339),
		},
	}

	if offset, err := ToInt(conversionOffsetField, datum[conversionOffsetField]); err == nil {
		base.ConversionOffset = offset
	} else {
		errs.Append(err)
	}

	if timezoneOffset, err := ToInt(timezoneOffsetField, datum[timezoneOffsetField]); err == nil {
		base.TimezoneOffset = timezoneOffset
	} else {
		errs.Append(err)
	}

	if clockDriftOffset, err := ToInt(clockDriftOffsetField, datum[clockDriftOffsetField]); err == nil {
		base.ClockDriftOffset = clockDriftOffset
	} else {
		errs.Append(err)
	}

	if deviceTime, err := ToString(deviceTimeField, datum[deviceTimeField]); err == nil {
		base.DeviceTime = deviceTime
	} else {
		errs.Append(err)
	}

	if annotations, err := ToArray(annotationsField, datum[annotationsField]); err == nil {
		base.Annotations = annotations
	} else {
		errs.Append(err)
	}

	validator.Struct(base, errs)

	return base
}

var validationFailureReasons = validate.ErrorReasons{
	timeStringTag:     timeStrValidationMsg,
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
	if timedatumect, ok := field.Interface().(time.Time); ok {
		return isTimeObjectValid(timedatumect)
	}
	return false
}

func isTimeObjectValid(timedatumect time.Time) bool {
	return !timedatumect.IsZero() && timedatumect.Before(time.Now())
}

func TimeStringValidator(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	if timeString, ok := field.Interface().(string); ok {
		return isTimeStringValid(timeString)
	}
	return false
}

func isTimeStringValid(timeString string) bool {
	var timedatumect time.Time
	timedatumect, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		timedatumect, err = time.Parse("2006-01-02T15:04:05", timeString)
		if err != nil {
			return false
		}
	}

	return isTimedatumectValid(timedatumect)
}

func ToString(fieldName string, data interface{}) (*string, *validate.Error) {
	if data == nil {
		return nil, nil
	}
	aString, ok := data.(*string)
	if !ok {
		return nil, validate.NewPointerError(fieldName, "Invalid type", "detail")
	}
	return aString, nil
}

func ToFloat64(fieldName string, data interface{}) (*float64, *validate.Error) {
	if data == nil {
		return nil, nil
	}
	theFloat, ok := data.(*float64)
	if !ok {
		return nil, validate.NewPointerError(fieldName, "Invalid type", "detail")
	}
	return theFloat, nil
}

func ToInt(fieldName string, data interface{}) (*int, *validate.Error) {
	if data == nil {
		return nil, nil
	}
	theInt, ok := data.(*int)
	if !ok {
		return nil, validate.NewPointerError(fieldName, "Invalid type", "detail")
	}
	return theInt, nil
}

func ToTime(fieldName string, data interface{}) (*time.Time, *validate.Error) {
	timeStr, ok := data.(string)
	if ok {
		theTime, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			//try this format also before we fail
			theTime, err = time.Parse("2006-01-02T15:04:05", timeStr)
		}

		c.err.AppendError(err)
		return theTime
	}
	c.err.AppendFieldError(fieldName, data)
	return time.Time{}
}

func ToArray(fieldName string, data interface{}) ([]interface{}, *validate.Error) {
	if data == nil {
		return nil
	}
	arrayData, ok := data.([]interface{})
	if !ok {
		c.err.AppendFieldError(fieldName, data)
		return nil
	}
	return arrayData
}
