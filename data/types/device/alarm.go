package device

import (
	"reflect"

	validator "gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(alarmTypeField.Tag, AlarmTypeValidator)
}

type Alarm struct {
	AlarmType *string `json:"alarmType" bson:"alarmType" valid:"devicealarmtype"`
	Status    *string `json:"status,omitempty" bson:"status,omitempty" valid:"-"`
	Base      `bson:",inline"`
}

var (
	alarmStatusField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "status"},
	}

	alarmTypeField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "alarmType"},
		Tag:        "devicealarmtype",
		Message:    "Must be one of low_insulin, no_insulin, low_power, no_power, occlusion, no_delivery, auto_off, over_limit, other",
		Allowed: types.Allowed{
			"low_insulin": true,
			"no_insulin":  true,
			"low_power":   true,
			"no_power":    true,
			"occlusion":   true,
			"no_delivery": true,
			"auto_off":    true,
			"over_limit":  true,
			"other":       true,
		},
	}
)

func (b Base) makeAlarm(datum types.Datum, errs validate.ErrorProcessing) *Alarm {
	Alarm := &Alarm{
		AlarmType: datum.ToString(alarmTypeField.Name, errs),
		Status:    datum.ToString(alarmStatusField.Name, errs),
		Base:      b,
	}
	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(Alarm, errs)
	return Alarm
}

func AlarmTypeValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	alarmType, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = alarmTypeField.Allowed[alarmType]
	return ok
}
