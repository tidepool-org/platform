package device

import (
	"reflect"

	validator "gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(subTypeField.Tag, SubTypeValidator)
	types.GetPlatformValidator().RegisterValidation(statusField.Tag, StatusValidator)
}

type Base struct {
	SubType    *string `json:"subType" bson:"subType" valid:"devicesubtype"`
	types.Base `bson:",inline"`
}

const Name = "deviceEvent"

var (
	subTypeField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "subType"},
		Tag:        "devicesubtype",
		Message:    "Must be one of alarm, calibration, status, prime, timeChange, reservoirChange",
		Allowed: types.Allowed{
			"alarm":           true,
			"calibration":     true,
			"status":          true,
			"prime":           true,
			"timeChange":      true,
			"reservoirChange": true,
		},
	}

	statusField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "status"},
		Tag:        "devicestatus",
		Message:    "Must be one of suspended, resumed",
		Allowed: types.Allowed{
			"suspended": true,
			"resumed":   true,
		},
	}

	failureReasons = validate.FailureReasons{
		"AlarmType":   validate.VaidationInfo{FieldName: alarmTypeField.Name, Message: alarmTypeField.Message},
		"Value":       validate.VaidationInfo{FieldName: types.BloodGlucoseValueField.Name, Message: types.BloodGlucoseValueField.Message},
		"Volume":      validate.VaidationInfo{FieldName: volumeField.Name, Message: volumeField.Message},
		"PrimeTarget": validate.VaidationInfo{FieldName: primeTargetField.Name, Message: primeTargetField.Message},
		"Reason":      validate.VaidationInfo{FieldName: reasonField.Name, Message: reasonField.Message},
		"Reasons":     validate.VaidationInfo{FieldName: timeChangeReasonsField.Name, Message: timeChangeReasonsField.Message},
		"Agent":       validate.VaidationInfo{FieldName: timeChangeAgentField.Name, Message: timeChangeAgentField.Message},
		"Status":      validate.VaidationInfo{FieldName: statusField.Name, Message: statusField.Message},
		"SubType":     validate.VaidationInfo{FieldName: subTypeField.Name, Message: subTypeField.Message},
	}
)

func Build(datum types.Datum, errs validate.ErrorProcessing) interface{} {

	base := Base{
		SubType: datum.ToString(subTypeField.Name, errs),
		Base:    types.BuildBase(datum, errs),
	}

	if base.SubType != nil {

		switch *base.SubType {
		case "alarm":
			return base.makeAlarm(datum, errs)
		case "calibration":
			return base.makeCalibration(datum, errs)
		case "status":
			return base.makeStatus(datum, errs)
		case "prime":
			return base.makePrime(datum, errs)
		case "timeChange":
			return base.makeTimeChange(datum, errs)
		case "reservoirChange":
			return base.makeReservoirChange(datum, errs)
		}
	}
	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(base, errs)
	return base
}

func StatusValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	status, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = statusField.Allowed[status]
	return ok
}

func SubTypeValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	subType, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = subTypeField.Allowed[subType]
	return ok
}
