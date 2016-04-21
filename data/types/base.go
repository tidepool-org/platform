package types

import (
	"fmt"
	"reflect"
	"time"

	validator "gopkg.in/bluesuncorp/validator.v8"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/validate"
)

func init() {
	GetPlatformValidator().RegisterValidation(baseTimezoneOffsetField.Tag, TimezoneOffsetValidator)
	GetPlatformValidator().RegisterValidation(basePayloadField.Tag, PayloadValidator)
	GetPlatformValidator().RegisterValidation(baseAnnotationsField.Tag, AnnotationsValidator)
}

type (
	Datum map[string]interface{}

	DatumArray []Datum

	Base struct {
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
	Internal struct {
		CreatedTime   string `json:"createdTime" bson:"createdTime" valid:"timestr"`
		GroupID       string `json:"-" bson:"_groupId" valid:"required"`
		ActiveFlag    bool   `json:"-" bson:"_active" valid:"required"`
		SchemaVersion int    `json:"-" bson:"_schemaVersion" valid:"required,min=0"`
		Version       int    `json:"-" bson:"_version,omitempty" valid:"-"`
	}
)

var (
	//InternalFields are what we only use internally in the service and don't wish to return to any clients
	InternalFields = []string{
		"_groupId",
		"_active",
		"_schemaVersion",
		"_version",
		"createdTime",
	}

	BaseUserIDField           = DatumField{Name: "userId"}
	BaseGroupIDField          = DatumField{Name: "groupId"}
	BaseInternalGroupIDField  = DatumField{Name: "_groupId"}
	BaseSubTypeField          = DatumField{Name: "subType"}
	baseUploadIDField         = DatumField{Name: "uploadId"}
	baseConversionOffsetField = DatumField{Name: "conversionOffset"}
	baseClockDriftOffsetField = DatumField{Name: "clockDriftOffset"}

	BaseTypeField = DatumFieldInformation{
		DatumField: &DatumField{Name: "type"},
		Tag:        "required",
		Message:    "This is a required field",
	}

	baseDeviceTimeField = DatumFieldInformation{
		DatumField: &DatumField{Name: "deviceTime"},
		Tag:        TimeStringField.Tag,
		Message:    TimeStringField.Message,
	}

	baseTimezoneOffsetField = IntDatumField{
		DatumField:      &DatumField{Name: "timezoneOffset"},
		Tag:             "timezoneoffset",
		Message:         "needs to be in minutes and >= -840 and <= 720",
		AllowedIntRange: &AllowedIntRange{LowerLimit: -840, UpperLimit: 720},
	}

	basePayloadField = DatumFieldInformation{
		DatumField: &DatumField{Name: "payload"},
		Tag:        "payload",
		Message:    "",
	}

	baseAnnotationsField = DatumFieldInformation{
		DatumField: &DatumField{Name: "annotations"},
		Tag:        "annotations",
		Message:    "",
	}

	BaseDeviceIDField = DatumFieldInformation{
		DatumField: &DatumField{Name: "deviceId"},
		Tag:        "required",
		Message:    "This is a required field",
	}

	failureReasons = validate.FailureReasons{
		"DeviceTime":     validate.ValidationInfo{FieldName: baseDeviceTimeField.Name, Message: baseDeviceTimeField.Message},
		"Time":           validate.ValidationInfo{FieldName: TimeStringField.Name, Message: TimeStringField.Message},
		"TimezoneOffset": validate.ValidationInfo{FieldName: baseTimezoneOffsetField.Name, Message: baseTimezoneOffsetField.Message},
		"Payload":        validate.ValidationInfo{FieldName: basePayloadField.Name, Message: basePayloadField.Message},
		"Annotations":    validate.ValidationInfo{FieldName: baseAnnotationsField.Name, Message: baseAnnotationsField.Message},
		"DeviceID":       validate.ValidationInfo{FieldName: BaseDeviceIDField.Name, Message: BaseDeviceIDField.Message},
		"Type":           validate.ValidationInfo{FieldName: BaseTypeField.Name, Message: BaseTypeField.Message},
	}
)

const (
	InvalidTypeTitle = "Invalid type"
	InvalidDataTitle = "Invalid data"

	invalidTypeDescription = "should be of type '%s'"
)

func BuildBase(datum Datum, errs validate.ErrorProcessing) Base {

	base := Base{
		_ID:              bson.NewObjectId(),
		ID:               bson.NewObjectId().Hex(),
		UserID:           datum.ToString(BaseUserIDField.Name, errs),
		DeviceID:         datum.ToString(BaseDeviceIDField.Name, errs),
		UploadID:         datum.ToString(baseUploadIDField.Name, errs),
		Time:             datum.ToString(TimeStringField.Name, errs),
		Type:             datum.ToString(BaseTypeField.Name, errs),
		Payload:          datum.ToObject(basePayloadField.Name, errs),
		ConversionOffset: datum.ToInt(baseConversionOffsetField.Name, errs),
		TimezoneOffset:   datum.ToInt(baseTimezoneOffsetField.Name, errs),
		ClockDriftOffset: datum.ToInt(baseClockDriftOffsetField.Name, errs),
		DeviceTime:       datum.ToString(baseDeviceTimeField.Name, errs),
		Annotations:      datum.ToArray(baseAnnotationsField.Name, errs),
		Internal: Internal{
			GroupID:       datum[BaseGroupIDField.Name].(string),
			ActiveFlag:    true,
			SchemaVersion: 1, //TODO: configured ??
			CreatedTime:   time.Now().Format(time.RFC3339),
		},
	}

	GetPlatformValidator().SetFailureReasons(failureReasons).Struct(base, errs)

	return base
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
	return offset >= baseTimezoneOffsetField.LowerLimit && offset <= baseTimezoneOffsetField.UpperLimit
}

func (d Datum) ToString(fieldName string, errs validate.ErrorProcessing) *string {
	if d[fieldName] == nil {
		return nil
	}
	aString, ok := d[fieldName].(string)
	if !ok {
		errs.AppendPointerError(fieldName, InvalidTypeTitle, fmt.Sprintf(invalidTypeDescription, "string"))
		return nil
	}
	return &aString
}

func (d Datum) ToFloat64(fieldName string, errs validate.ErrorProcessing) *float64 {
	if d[fieldName] == nil {
		return nil
	}
	theFloat, ok := d[fieldName].(float64)
	if !ok {
		errs.AppendPointerError(fieldName, InvalidTypeTitle, fmt.Sprintf(invalidTypeDescription, "float"))
		return nil
	}
	return &theFloat
}

func (d Datum) ToInt(fieldName string, errs validate.ErrorProcessing) *int {
	if d[fieldName] == nil {
		return nil
	}
	theInt, _ := d[fieldName].(int)
	//TODO:
	/*if !ok {
		return 0
		appendInvalidTypeError(errs, fieldName, "integer")
		return 0
	}*/
	return &theInt
}

func (d Datum) ToTime(fieldName string, errs validate.ErrorProcessing) *time.Time {

	timeStr, ok := d[fieldName].(string)
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

func (d Datum) ToArray(fieldName string, errs validate.ErrorProcessing) *[]interface{} {
	if d[fieldName] == nil {
		return nil
	}
	arrayData, ok := d[fieldName].([]interface{})
	if !ok {
		errs.AppendPointerError(fieldName, InvalidTypeTitle, fmt.Sprintf(invalidTypeDescription, "array"))
		return nil
	}
	return &arrayData
}

func (d Datum) ToStringArray(fieldName string, errs validate.ErrorProcessing) *[]string {
	if d[fieldName] == nil {
		return nil
	}
	arrayData, ok := d[fieldName].([]string)
	if !ok {
		errs.AppendPointerError(fieldName, InvalidTypeTitle, fmt.Sprintf(invalidTypeDescription, "string array"))
		return nil
	}
	return &arrayData
}

func (d Datum) ToObject(fieldName string, errs validate.ErrorProcessing) *interface{} {
	if d[fieldName] == nil {
		return nil
	}
	objectData, ok := d[fieldName].(interface{})
	if !ok {
		errs.AppendPointerError(fieldName, InvalidTypeTitle, fmt.Sprintf(invalidTypeDescription, "object"))
		return nil
	}
	return &objectData
}
