package upload

import (
	"reflect"

	validator "gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(deviceTagsField.Tag, DeviceTagsValidator)
	types.GetPlatformValidator().RegisterValidation(timeProcessingField.Tag, TimeProcessingValidator)
	types.GetPlatformValidator().RegisterValidation(deviceManufacturersField.Tag, DeviceManufacturersValidator)
}

type Upload struct {
	UploadID            *string     `json:"uploadId" bson:"uploadId" valid:"gt=10"`
	UploadUserID        *string     `json:"byUser" bson:"byUser" valid:"gte=10"`
	Version             *string     `json:"version" bson:"version" valid:"gte=5"`
	ComputerTime        *string     `json:"computerTime" bson:"computerTime" valid:"timestr"`
	DeviceTags          *[]string   `json:"deviceTags" bson:"deviceTags" valid:"uploaddevicetags"`
	DeviceManufacturers *[]string   `json:"deviceManufacturers" bson:"deviceManufacturers" valid:"uploaddevicemanufacturers"`
	DeviceModel         *string     `json:"deviceModel" bson:"deviceModel" valid:"gte=1"`
	DeviceSerialNumber  *string     `json:"deviceSerialNumber" bson:"deviceSerialNumber" valid:"gte=10"`
	TimeProcessing      *string     `json:"timeProcessing" bson:"timeProcessing" valid:"uploadtimeprocessing"`
	DataState           *string     `json:"dataState" bson:"dataState"`
	Deduplicator        interface{} `json:"deduplicator" bson:"deduplicator"`
	types.Base          `bson:",inline"`
}

const Name = "upload"

var (
	deviceTagsField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "deviceTags"},
		Tag:        "uploaddevicetags",
		Message:    "Must be one of insulin-pump, cgm, bgm",
		Allowed: types.Allowed{
			"insulin-pump": true,
			"cgm":          true,
			"bgm":          true,
		},
	}

	deviceManufacturersField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "deviceManufacturers"},
		Tag:        "uploaddevicemanufacturers",
		Message:    "Must contain at least one manufacturer name",
	}

	timeProcessingField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "timeProcessing"},
		Tag:        "uploadtimeprocessing",
		Message:    "Must be one of across-the-board-timezone, utc-bootstrapping, none",
		Allowed: types.Allowed{
			"across-the-board-timezone": true,
			"utc-bootstrapping":         true,
			"none":                      true,
		},
	}

	computerTimeField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "computerTime"},
		Tag:        "timestr",
		Message:    types.TimeStringField.Message,
	}

	uploadIDField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "uploadId"},
		Tag:        "gte",
		Message:    "This is a required field need needs to be 10+ characters in length",
	}

	uploadUserIDField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "byUser"},
		Tag:        "gte",
		Message:    "This is a required field need needs to be 10+ characters in length",
	}

	deviceModelField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "deviceModel"},
		Tag:        "gte",
		Message:    "This is a required field need needs to be 10+ characters in length",
	}

	deviceSerialNumberField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "deviceSerialNumber"},
		Tag:        "gte",
		Message:    "This is a required field need needs to be 10+ characters in length",
	}

	versionField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "version"},
		Tag:        "gte",
		Message:    "This is a required field need needs to be 10+ characters in length",
	}

	failureReasons = validate.FailureReasons{
		"DeviceTags":          validate.ValidationInfo{FieldName: deviceTagsField.Name, Message: deviceTagsField.Message},
		"TimeProcessing":      validate.ValidationInfo{FieldName: timeProcessingField.Name, Message: timeProcessingField.Message},
		"DeviceManufacturers": validate.ValidationInfo{FieldName: deviceManufacturersField.Name, Message: deviceManufacturersField.Message},
		"ComputerTime":        validate.ValidationInfo{FieldName: computerTimeField.Name, Message: computerTimeField.Message},
		"UploadID":            validate.ValidationInfo{FieldName: uploadIDField.Name, Message: uploadIDField.Message},
		"UploadUserID":        validate.ValidationInfo{FieldName: uploadUserIDField.Name, Message: uploadUserIDField.Message},
		"DeviceModel":         validate.ValidationInfo{FieldName: deviceModelField.Name, Message: deviceModelField.Message},
		"DeviceSerialNumber":  validate.ValidationInfo{FieldName: deviceSerialNumberField.Name, Message: deviceSerialNumberField.Message},
		"Version":             validate.ValidationInfo{FieldName: versionField.Name, Message: versionField.Message},
	}
)

func Build(datum types.Datum, errs validate.ErrorProcessing) *Upload {

	record := &Upload{
		UploadID:            datum.ToString(uploadIDField.Name, errs),
		ComputerTime:        datum.ToString(computerTimeField.Name, errs),
		UploadUserID:        datum.ToString(uploadUserIDField.Name, errs),
		Version:             datum.ToString(versionField.Name, errs),
		TimeProcessing:      datum.ToString(timeProcessingField.Name, errs),
		DeviceModel:         datum.ToString(deviceModelField.Name, errs),
		DeviceManufacturers: datum.ToStringArray(deviceManufacturersField.Name, errs),
		DeviceTags:          datum.ToStringArray(deviceTagsField.Name, errs),
		DeviceSerialNumber:  datum.ToString(deviceSerialNumberField.Name, errs),
		Base:                types.BuildBase(datum, errs),
	}

	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(record, errs)

	return record
}

func TimeProcessingValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	procesingType, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = timeProcessingField.Allowed[procesingType]
	return ok
}

func DeviceTagsValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	tags, ok := field.Interface().([]string)

	if !ok {
		return false
	}
	if len(tags) == 0 {
		return false
	}
	for i := range tags {
		_, ok = deviceTagsField.Allowed[tags[i]]
		if ok == false {
			break
		}
	}
	return ok
}

func DeviceManufacturersValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	deviceManufacturersField, ok := field.Interface().([]string)
	if !ok {
		return false
	}
	return len(deviceManufacturersField) > 0
}
