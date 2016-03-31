package data

import (
	"fmt"
	"reflect"
	"time"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
	"github.com/tidepool-org/platform/Godeps/_workspace/src/labix.org/v2/mgo/bson"

	"github.com/tidepool-org/platform/validate"
)

func init() {
	getPlatformValidator().RegisterValidation(timeStringTag, PastTimeStringValidator)
	getPlatformValidator().RegisterValidation(timeObjectTag, PastTimeObjectValidator)
	getPlatformValidator().RegisterValidation(timeZoneOffsetTag, TimezoneOffsetValidator)
	getPlatformValidator().RegisterValidation(payloadTag, PayloadValidator)
	getPlatformValidator().RegisterValidation(annotationsTag, AnnotationsValidator)
}

type Base struct {
	//required data
	_ID      bson.ObjectId `bson:"_id" valid:"mongo,required"`
	ID       string        `json:"id" bson:"id" valid:"required"`
	UserID   *string       `json:"userId" bson:"userId" valid:"required"`
	DeviceID *string       `json:"deviceId" bson:"deviceId" valid:"required"`
	Time     *string       `json:"time" bson:"time" valid:"timestr"`
	Type     *string       `json:"type" bson:"type" valid:"required"`
	UploadID *string       `json:"uploadId" bson:"uploadId" valid:"-"`

	//optional data
	DeviceTime       *string        `json:"deviceTime,omitempty" bson:"deviceTime,omitempty" valid:"omitempty,timestr"`
	TimezoneOffset   *int           `json:"timezoneOffset,omitempty" bson:"timezoneOffset,omitempty" valid:"omitempty,timezoneoffset"`
	ConversionOffset *int           `json:"conversionOffset,omitempty" bson:"conversionOffset,omitempty" valid:"omitempty,required"`
	ClockDriftOffset *int           `json:"clockDriftOffset,omitempty" bson:"clockDriftOffset,omitempty" valid:"omitempty,required"`
	Payload          *interface{}   `json:"payload,omitempty" bson:"payload,omitempty" valid:"omitempty,payload"`
	Annotations      *[]interface{} `json:"annotations,omitempty" bson:"annotations,omitempty" valid:"omitempty,annotations"`

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
	//InternalFields are what we only use internally in the service and don't wish to return to any clients
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

	invalidTypeDescription = "should be of type '%s'"
)

func BuildBase(datum Datum, errs validate.ErrorProcessing) Base {

	base := Base{
		_ID:              bson.NewObjectId(),
		ID:               bson.NewObjectId().Hex(),
		UserID:           ToString(UserIDField, datum[UserIDField], errs),
		DeviceID:         ToString(deviceIDField, datum[deviceIDField], errs),
		UploadID:         ToString(uploadIDField, datum[uploadIDField], errs),
		Time:             ToString(timeField, datum[timeField], errs),
		Type:             ToString(typeField, datum[typeField], errs),
		Payload:          ToObject(payloadField, datum[payloadField], errs),
		ConversionOffset: ToInt(conversionOffsetField, datum[conversionOffsetField], errs),
		TimezoneOffset:   ToInt(timezoneOffsetField, datum[timezoneOffsetField], errs),
		ClockDriftOffset: ToInt(clockDriftOffsetField, datum[clockDriftOffsetField], errs),
		DeviceTime:       ToString(deviceTimeField, datum[deviceTimeField], errs),
		Annotations:      ToArray(annotationsField, datum[annotationsField], errs),
		Internal: Internal{
			GroupID:       datum[GroupIDField].(string),
			ActiveFlag:    true,
			SchemaVersion: 1, //TODO: configured ??
			CreatedTime:   time.Now().Format(time.RFC3339),
		},
	}

	getPlatformValidator().SetErrorReasons(validationFailureReasons).Struct(base, errs)

	return base
}

var validationFailureReasons = validate.ErrorReasons{
	timeStringTag:     timeStrValidationMsg,
	timeZoneOffsetTag: fmt.Sprintf("TimezoneOffset needs to be in minutes and greater than %d and less than %d", tzValidationLowerLimit, tzValidationUpperLimit),
}

func PayloadValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	//TODO: a place holder for more through validation
	return true
}

func AnnotationsValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	//TODO: a place holder for more through validation
	return true
}

func TimezoneOffsetValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	offset, ok := field.Interface().(int)
	if !ok {
		return false
	}
	//TODO: needs to be confirmed that this is all we should validate
	return offset >= tzValidationLowerLimit && offset <= tzValidationUpperLimit
}

func PastTimeObjectValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	timeObject, ok := field.Interface().(time.Time)
	if !ok {
		return false
	}
	return isTimeObjectValid(timeObject)
}

func isTimeObjectValid(timeObject time.Time) bool {
	return !timeObject.IsZero() && timeObject.Before(time.Now())
}

func PastTimeStringValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	timeString, ok := field.Interface().(string)
	if !ok {
		return false
	}
	return isTimeStringValid(timeString)
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

func ToString(fieldName string, data interface{}, errs validate.ErrorProcessing) *string {
	if data == nil {
		return nil
	}
	aString, ok := data.(string)
	if !ok {
		errs.AppendPointerError(fieldName, InvalidTypeTitle, fmt.Sprintf(invalidTypeDescription, "string"))
		return nil
	}
	return &aString
}

func ToFloat64(fieldName string, data interface{}, errs validate.ErrorProcessing) *float64 {
	if data == nil {
		return nil
	}
	theFloat, ok := data.(float64)
	if !ok {
		errs.AppendPointerError(fieldName, InvalidTypeTitle, fmt.Sprintf(invalidTypeDescription, "float"))
		return nil
	}
	return &theFloat
}

func ToInt(fieldName string, data interface{}, errs validate.ErrorProcessing) *int {
	if data == nil {
		return nil
	}
	theInt, _ := data.(int)
	//TODO:
	/*if !ok {
		return 0
		appendInvalidTypeError(errs, fieldName, "integer")
		return 0
	}*/
	return &theInt
}

func ToTime(fieldName string, data interface{}, errs validate.ErrorProcessing) *time.Time {

	timeStr, ok := data.(string)
	if !ok {
		errs.AppendPointerError(fieldName, InvalidTypeTitle, fmt.Sprintf(invalidTypeDescription, "string"))
		return nil
	}
	theTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		//try this format also before we fail
		theTime, err = time.Parse("2006-01-02T15:04:05", timeStr)
		if err != nil {
			errs.AppendPointerError(fieldName, InvalidTypeTitle, fmt.Sprintf(invalidTypeDescription, "string"))
			return nil
		}
	}
	return &theTime
}

func ToArray(fieldName string, data interface{}, errs validate.ErrorProcessing) *[]interface{} {
	if data == nil {
		return nil
	}
	arrayData, ok := data.([]interface{})
	if !ok {
		errs.AppendPointerError(fieldName, InvalidTypeTitle, fmt.Sprintf(invalidTypeDescription, "array"))
		return nil
	}
	return &arrayData
}

func ToObject(fieldName string, data interface{}, errs validate.ErrorProcessing) *interface{} {
	if data == nil {
		return nil
	}
	objectData, ok := data.(interface{})
	if !ok {
		errs.AppendPointerError(fieldName, InvalidTypeTitle, fmt.Sprintf(invalidTypeDescription, "object"))
		return nil
	}
	return &objectData
}
