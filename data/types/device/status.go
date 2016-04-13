package device

import (
	"reflect"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(reasonField.Tag, ReasonValidator)
}

type Status struct {
	Status *string      `json:"status" bson:"status" valid:"devicestatus"`
	Reason *interface{} `json:"reason" bson:"reason" valid:"devicereason"`
	Base   `bson:",inline"`
}

var (
	reasonField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "reason"},
		Tag:        "devicereason",
		Message:    "Must be one of manual, automatic",
		Allowed: types.Allowed{
			"manual":    true,
			"automatic": true,
		},
	}
)

func (b Base) makeStatus(datum types.Datum, errs validate.ErrorProcessing) *Status {
	status := &Status{
		Status: datum.ToString(statusField.Name, errs),
		Reason: datum.ToObject(reasonField.Name, errs),
		Base:   b,
	}
	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(status, errs)
	return status
}

func ReasonValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {

	return true
	/*
		TODO:
		reason, ok := field.Interface().(map[string]string)

		if !ok {
			return false
		}

		for _, val := range reason {
			_, ok = reasonField.Allowed[val]
		}
		return ok*/
}
