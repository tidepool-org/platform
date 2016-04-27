package basal

import (
	"reflect"

	validator "gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(durationField.Tag, DurationValidator)
	types.GetPlatformValidator().RegisterValidation(deliveryTypeField.Tag, DeliveryTypeValidator)
	types.GetPlatformValidator().RegisterValidation(rateField.Tag, RateValidator)
}

type Base struct {
	DeliveryType *string `json:"deliveryType" bson:"deliveryType" valid:"basaldeliverytype"`
	types.Base   `bson:",inline"`
}

type Suppressed struct {
	DeliveryType *string  `json:"deliveryType" bson:"deliveryType" valid:"basaldeliverytype"`
	ScheduleName *string  `json:"scheduleName" bson:"scheduleName" valid:"omitempty,required"`
	Rate         *float64 `json:"rate" bson:"rate" valid:"basalrate"`
}

const Name = "basal"

var (
	deliveryTypeField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "deliveryType"},
		Tag:        "basaldeliverytype",
		Message:    "Must be one of scheduled, suspend, temp",
		Allowed:    types.Allowed{"scheduled": true, "suspend": true, "temp": true},
	}

	durationField = types.IntDatumField{
		DatumField:      &types.DatumField{Name: "duration"},
		Tag:             "basalduration",
		Message:         "Must be >= 0 and <= 432000000",
		AllowedIntRange: &types.AllowedIntRange{LowerLimit: 0, UpperLimit: 432000000},
	}

	rateField = types.FloatDatumField{
		DatumField:        &types.DatumField{Name: "rate"},
		Tag:               "basalrate",
		Message:           "Must be >= 0.0 and <= 20.0",
		AllowedFloatRange: &types.AllowedFloatRange{LowerLimit: 0.0, UpperLimit: 20.0},
	}

	failureReasons = validate.FailureReasons{
		"DeliveryType": validate.ValidationInfo{FieldName: deliveryTypeField.Name, Message: deliveryTypeField.Message},
		"Rate":         validate.ValidationInfo{FieldName: rateField.Name, Message: rateField.Message},
		"Duration":     validate.ValidationInfo{FieldName: durationField.Name, Message: durationField.Message},
		"TempDuration": validate.ValidationInfo{FieldName: tempDurationField.Name, Message: tempDurationField.Message},
		"Percent":      validate.ValidationInfo{FieldName: percentField.Name, Message: percentField.Message},
	}
)

func makeSuppressed(datum types.Datum, errs validate.ErrorProcessing) *Suppressed {
	return &Suppressed{
		DeliveryType: datum.ToString(deliveryTypeField.Name, errs),
		ScheduleName: datum.ToString(scheduleNameField.Name, errs),
		Rate:         datum.ToFloat64(rateField.Name, errs),
	}
}

func Build(datum types.Datum, errs validate.ErrorProcessing) interface{} {

	base := &Base{
		DeliveryType: datum.ToString(deliveryTypeField.Name, errs),
		Base:         types.BuildBase(datum, errs),
	}

	if base.DeliveryType != nil {

		switch *base.DeliveryType {
		case "scheduled":
			return base.makeScheduled(datum, errs)
		case "suspend":
			return base.makeSuspend(datum, errs)
		case "temp":
			return base.makeTemporary(datum, errs)
		}
	}
	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(base, errs)
	return base
}

func RateValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	rate, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return rate >= rateField.LowerLimit && rate <= rateField.UpperLimit
}

func DurationValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	duration, ok := field.Interface().(int)
	if !ok {
		return false
	}
	return duration >= durationField.LowerLimit && duration <= durationField.UpperLimit
}

func DeliveryTypeValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	deliveryType, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = deliveryTypeField.Allowed[deliveryType]
	return ok
}
