package data

import (
	"reflect"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/validate"
)

func init() {
	getPlatformValidator().RegisterValidation(basalRateField.Tag, BasalRateValidator)
	getPlatformValidator().RegisterValidation(basalDurationField.Tag, BasalDurationValidator)
	getPlatformValidator().RegisterValidation(basalDeliveryTypeField.Tag, BasalDeliveryTypeValidator)
	getPlatformValidator().RegisterValidation(basalInsulinField.Tag, BasalInsulinValidator)
	getPlatformValidator().RegisterValidation(basalValueField.Tag, BasalValueValidator)
}

type Basal struct {
	DeliveryType *string          `json:"deliveryType" bson:"deliveryType" valid:"basaldeliverytype"`
	ScheduleName *string          `json:"scheduleName" bson:"scheduleName" valid:"omitempty,required"`
	Rate         *float64         `json:"rate,omitempty" bson:"rate,omitempty" valid:"omitempty,basalrate"`
	Duration     *int             `json:"duration,omitempty" bson:"duration,omitempty" valid:"omitempty,basalduration"`
	Insulin      *string          `json:"insulin,omitempty" bson:"insulin,omitempty" valid:"omitempty,basalinsulin"`
	Value        *int             `json:"value,omitempty" bson:"value,omitempty" valid:"omitempty,basalvalue"`
	Suppressed   *SuppressedBasal `json:"suppressed,omitempty" bson:"suppressed,omitempty" valid:"omitempty,required"`
	Base         `bson:",inline"`
}

type SuppressedBasal struct {
	Type         *string  `json:"type" bson:"type" valid:"required"`
	DeliveryType *string  `json:"deliveryType" bson:"deliveryType" valid:"basaldeliverytype"`
	ScheduleName *string  `json:"scheduleName" bson:"scheduleName" valid:"omitempty,required"`
	Rate         *float64 `json:"rate" bson:"rate" valid:"omitempty,basalrate"`
}

const BasalName = "basal"

var (
	basalDeliveryTypeField = TypesDatumField{
		DatumField:   &DatumField{Name: "deliveryType"},
		Tag:          "basaldeliverytype",
		Message:      "Must be one of injected, scheduled, suspend, temp",
		AllowedTypes: AllowedTypes{"injected": true, "scheduled": true, "suspend": true, "temp": true},
	}

	basalRateField = FloatDatumField{
		DatumField:        &DatumField{Name: "rate"},
		Tag:               "basalrate",
		Message:           "Must be greater than 0.0",
		AllowedFloatRange: &AllowedFloatRange{LowerLimit: 0.0},
	}

	basalScheduleNameField = DatumField{Name: "scheduleName"}

	basalInsulinField = TypesDatumField{
		DatumField:   &DatumField{Name: "insulin"},
		Tag:          "basalinsulin",
		Message:      "Must be one of levemir, lantus",
		AllowedTypes: AllowedTypes{"levemir": true, "lantus": true},
	}

	basalDurationField = IntDatumField{
		DatumField:      &DatumField{Name: "duration"},
		Tag:             "basalduration",
		Message:         "Must be greater than 0",
		AllowedIntRange: &AllowedIntRange{LowerLimit: 0},
	}

	basalValueField = IntDatumField{
		DatumField:      &DatumField{Name: "value"},
		Tag:             "basalvalue",
		Message:         "Must be greater than 0",
		AllowedIntRange: &AllowedIntRange{LowerLimit: 0},
	}

	basalFailureReasons = validate.ErrorReasons{
		basalDeliveryTypeField.Tag: basalDeliveryTypeField.Message,
		basalRateField.Tag:         basalRateField.Message,
		basalDurationField.Tag:     basalDurationField.Message,
		basalValueField.Tag:        basalValueField.Message,
		basalInsulinField.Tag:      basalInsulinField.Message,
	}
)

//BuildBasal will build a Basal record
func BuildBasal(datum Datum, errs validate.ErrorProcessing) *Basal {

	basal := &Basal{
		ScheduleName: datum.ToString(basalScheduleNameField.Name, errs),
		DeliveryType: datum.ToString(basalDeliveryTypeField.Name, errs),
		Rate:         datum.ToFloat64(basalRateField.Name, errs),
		Duration:     datum.ToInt(basalDurationField.Name, errs),
		Insulin:      datum.ToString(basalInsulinField.Name, errs),
		Base:         BuildBase(datum, errs),
	}

	getPlatformValidator().SetErrorReasons(basalFailureReasons).Struct(basal, errs)

	return basal
}

func BasalRateValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	rate, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return rate > basalRateField.LowerLimit
}

func BasalDurationValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	duration, ok := field.Interface().(int)
	if !ok {
		return false
	}
	return duration > basalDurationField.LowerLimit
}

func BasalDeliveryTypeValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	deliveryType, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = basalDeliveryTypeField.AllowedTypes[deliveryType]
	return ok
}

func BasalInsulinValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	insulin, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = basalInsulinField.AllowedTypes[insulin]
	return ok
}

func BasalValueValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	value, ok := field.Interface().(int)
	if !ok {
		return false
	}
	return value > basalValueField.LowerLimit
}
