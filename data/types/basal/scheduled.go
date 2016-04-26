package basal

import (
	"reflect"

	validator "gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(rateField.Tag, RateValidator)
}

type Scheduled struct {
	ScheduleName *string  `json:"scheduleName" bson:"scheduleName" valid:"omitempty,required"`
	Rate         *float64 `json:"rate" bson:"rate" valid:"required,basalrate"`
	Base         `bson:",inline"`
}

var (
	rateField = types.FloatDatumField{
		DatumField:        &types.DatumField{Name: "rate"},
		Tag:               "basalrate",
		Message:           "Must be  >= 0.0 and <= 20.0",
		AllowedFloatRange: &types.AllowedFloatRange{LowerLimit: 0.0, UpperLimit: 20.0},
	}

	scheduleNameField = types.DatumField{Name: "scheduleName"}
)

func (b Base) makeScheduled(datum types.Datum, errs validate.ErrorProcessing) *Scheduled {
	scheduled := &Scheduled{
		ScheduleName: datum.ToString(scheduleNameField.Name, errs),
		Rate:         datum.ToFloat64(rateField.Name, errs),
		Base:         b,
	}
	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(scheduled, errs)
	return scheduled
}

func RateValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	rate, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return rate >= rateField.LowerLimit && rate <= rateField.UpperLimit
}
