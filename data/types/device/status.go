package device

import (
	"reflect"

	validator "gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(statusReasonField.Tag, StatusReasonValidator)
	types.GetPlatformValidator().RegisterValidation(statusDurationField.Tag, StatusDurationValidator)
}

type Status struct {
	Status *string                `json:"status" bson:"status" valid:"devicestatus"`
	Reason map[string]interface{} `json:"reason" bson:"reason" valid:"devicereason"`
	//TODO: this should become required for the platform but is currently optional
	Duration *int `json:"duration,omitempty" bson:"duration,omitempty" valid:"omitempty,devicestatusduration"`
	Base     `bson:",inline"`
}

var (
	statusReasonField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "reason"},
		Tag:        "devicereason",
		Message:    "Must be one of manual, automatic",
		Allowed: types.Allowed{
			"manual":    true,
			"automatic": true,
		},
	}

	statusDurationField = types.IntDatumField{
		DatumField:      &types.DatumField{Name: "duration"},
		Tag:             "devicestatusduration",
		Message:         "Must be one of manual, automatic",
		AllowedIntRange: &types.AllowedIntRange{LowerLimit: 0},
	}
)

func (b Base) makeStatus(datum types.Datum, errs validate.ErrorProcessing) *Status {
	status := &Status{
		Status:   datum.ToString(statusField.Name, errs),
		Reason:   datum.ToMap(statusReasonField.Name, errs),
		Duration: datum.ToInt(statusDurationField.Name, errs),
		Base:     b,
	}
	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(status, errs)
	return status
}

func StatusReasonValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	reason, ok := field.Interface().(map[string]interface{})
	if !ok || reason == nil {
		return false
	}

	for _, v := range reason {
		_, ok = statusReasonField.Allowed[v.(string)]
		if !ok {
			return false
		}
	}

	return true

}

func StatusDurationValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	duration, ok := field.Interface().(int)
	if !ok {
		return false
	}

	return duration >= statusDurationField.LowerLimit

}
