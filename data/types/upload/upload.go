package upload

import (
	"fmt"
	"reflect"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(deviceTagsField.Tag, DeviceTagsValidator)
	types.GetPlatformValidator().RegisterValidation(timeProcessingField.Tag, TimeProcessingValidator)
	types.GetPlatformValidator().RegisterValidation(deviceManufacturersField.Tag, DeviceManufacturersValidator)
}

type Record struct {
	UploadID            *string   `json:"uploadId" bson:"uploadId" valid:"required,gt=10"`
	UploadUserID        *string   `json:"byUser" bson:"byUser" valid:"required,gt=10"`
	Version             *string   `json:"version" bson:"version" valid:"required,gt=10"`
	ComputerTime        *string   `json:"computerTime" bson:"computerTime" valid:"timestr"`
	DeviceID            *string   `json:"deviceId" bson:"deviceId" valid:"required,gt=10"`
	DeviceTags          *[]string `json:"deviceTags" bson:"deviceTags" valid:"uploaddevicetags"`
	DeviceManufacturers *[]string `json:"deviceManufacturers" bson:"deviceManufacturers" valid:"uploaddevicemanufacturers"`
	DeviceModel         *string   `json:"deviceModel" bson:"deviceModel" valid:"required,gt=10"`
	DeviceSerialNumber  *string   `json:"deviceSerialNumber" bson:"deviceSerialNumber" valid:"required,gt=10"`
	TimeProcessing      *string   `json:"timeProcessing" bson:"timeProcessing" valid:"uploadtimeprocessing"`
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

	computerTimeField       = types.DatumField{Name: "computerTime"}
	uploadIDField           = types.DatumField{Name: "uploadId"}
	uploadUserIDField       = types.DatumField{Name: "byUser"}
	deviceIDField           = types.DatumField{Name: "deviceId"}
	deviceModelField        = types.DatumField{Name: "deviceModel"}
	deviceSerialNumberField = types.DatumField{Name: "deviceSerialNumber"}
	versionField            = types.DatumField{Name: "version"}

	failureReasons = validate.ErrorReasons{
		deviceTagsField.Tag:          deviceTagsField.Message,
		timeProcessingField.Tag:      timeProcessingField.Message,
		deviceManufacturersField.Tag: deviceManufacturersField.Message,
	}
)

func Build(datum types.Datum, errs validate.ErrorProcessing) *Record {

	record := &Record{
		UploadID:            datum.ToString(uploadIDField.Name, errs),
		ComputerTime:        datum.ToString(computerTimeField.Name, errs),
		UploadUserID:        datum.ToString(uploadUserIDField.Name, errs),
		Version:             datum.ToString(versionField.Name, errs),
		TimeProcessing:      datum.ToString(timeProcessingField.Name, errs),
		DeviceModel:         datum.ToString(deviceModelField.Name, errs),
		DeviceManufacturers: datum.ToStringArray(deviceManufacturersField.Name, errs),
		DeviceTags:          datum.ToStringArray(deviceTagsField.Name, errs),
		DeviceSerialNumber:  datum.ToString(deviceSerialNumberField.Name, errs),
		DeviceID:            datum.ToString(deviceIDField.Name, errs),
		Base:                types.BuildBase(datum, errs),
	}

	types.GetPlatformValidator().Struct(record, errs)

	//types.GetPlatformValidator().SetErrorReasons(failureReasons).Struct(record, errs)

	if errs.HasErrors() {
		fmt.Println("## Errors ## ", errs.Errors)
	}

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
