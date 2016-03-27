package data

import (
	"fmt"
	"reflect"
	"strings"

	valid "github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
	"github.com/tidepool-org/platform/validate"
)

func init() {
	validator.RegisterValidation(rateTag, BasalRateValidator)
	validator.RegisterValidation(durationTag, BasalDurationValidator)
	validator.RegisterValidation(deliveryTypeTag, BasalDeliveryTypeValidator)
	validator.RegisterValidation(insulinTag, BasalInsulinValidator)
	validator.RegisterValidation(valueTag, BasalValueValidator)
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

	validator.SetErrorReasons(basalFailureReasons).Struct(basal, errs)

	return basal
}

func BasalRateValidator(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	if rate, ok := field.Interface().(float64); ok {
		if rate > rateValidationLowerLimit {
			return true
		}
	}
	return false
}

func BasalDurationValidator(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	if duration, ok := field.Interface().(int); ok {
		if duration > durationValidationLowerLimit {
			return true
		}
	}
	return false
}

func BasalDeliveryTypeValidator(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	if deliveryType, ok := field.Interface().(string); ok {
		if _, ok = allowedDeliveryTypes[strings.ToLower(deliveryType)]; ok {
			return true
		}
	}
	return false
}

func BasalInsulinValidator(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	if insulin, ok := field.Interface().(string); ok {
		if _, ok = allowedInsulins[strings.ToLower(insulin)]; ok {
			return true
		}
	}
	return false
}

func BasalValueValidator(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	if value, ok := field.Interface().(int); ok {
		if value > valueValidationLowerLimit {
			return true
		}
	}
	return false
}
