package data

import (
	"fmt"
	"reflect"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/validate"
)

func init() {
	getPlatformValidator().RegisterValidation(rateTag, BasalRateValidator)
	getPlatformValidator().RegisterValidation(durationTag, BasalDurationValidator)
	getPlatformValidator().RegisterValidation(deliveryTypeTag, BasalDeliveryTypeValidator)
	getPlatformValidator().RegisterValidation(insulinTag, BasalInsulinValidator)
	getPlatformValidator().RegisterValidation(valueTag, BasalValueValidator)
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

const (
	BasalName = "basal"

	deliveryTypeField = "deliveryType"
	scheduleNameField = "scheduleName"
	insulinField      = "insulin"
	valueField        = "value"
	rateField         = "rate"
	durationField     = "duration"

	deliveryTypeTag validate.ValidationTag = "basaldeliverytype"
	rateTag         validate.ValidationTag = "basalrate"
	durationTag     validate.ValidationTag = "basalduration"
	insulinTag      validate.ValidationTag = "basalinsulin"
	valueTag        validate.ValidationTag = "basalvalue"

	injectedDelivery  = "injected"
	scheduledDelivery = "scheduled"
	suspendDelivery   = "suspend"
	tempDelivery      = "temp"

	levemirInsulin = "levemir"
	lantusInsulin  = "lantus"

	rateValidationLowerLimit     = 0.0
	durationValidationLowerLimit = 0
	valueValidationLowerLimit    = 0
)

var (
	allowedDeliveryTypes = map[string]string{injectedDelivery: injectedDelivery, scheduledDelivery: scheduledDelivery, suspendDelivery: suspendDelivery, tempDelivery: tempDelivery}
	allowedInsulins      = map[string]string{levemirInsulin: levemirInsulin, lantusInsulin: lantusInsulin}

	basalFailureReasons = validate.ErrorReasons{
		deliveryTypeTag: fmt.Sprintf("Must be one of %s,%s,%s,%s", injectedDelivery, scheduledDelivery, suspendDelivery, tempDelivery),
		rateTag:         fmt.Sprintf("Must be greater than %.1f", rateValidationLowerLimit),
		durationTag:     fmt.Sprintf("Must be greater than %d", durationValidationLowerLimit),
		valueTag:        fmt.Sprintf("Must be greater than %d", valueValidationLowerLimit),
		insulinTag:      fmt.Sprintf("Must be one of %s,%s", levemirInsulin, lantusInsulin),
	}
)

//BuildBasal will build a Basal record
func BuildBasal(datum Datum, errs validate.ErrorProcessing) *Basal {

	basal := &Basal{
		ScheduleName: ToString(scheduleNameField, datum[scheduleNameField], errs),
		DeliveryType: ToString(deliveryTypeField, datum[deliveryTypeField], errs),
		Rate:         ToFloat64(rateField, datum[rateField], errs),
		Duration:     ToInt(durationField, datum[durationField], errs),
		Insulin:      ToString(insulinField, datum[insulinField], errs),
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
	return rate > rateValidationLowerLimit
}

func BasalDurationValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	duration, ok := field.Interface().(int)
	if !ok {
		return false
	}
	return duration > durationValidationLowerLimit
}

func BasalDeliveryTypeValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	deliveryType, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = allowedDeliveryTypes[deliveryType]
	return ok
}

func BasalInsulinValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	insulin, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = allowedInsulins[insulin]
	return ok
}

func BasalValueValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	value, ok := field.Interface().(int)
	if !ok {
		return false
	}
	return value > valueValidationLowerLimit
}
