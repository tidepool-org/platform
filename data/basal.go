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

//Basal represents a basal device data record
type Basal struct {
	DeliveryType string          `json:"deliveryType" bson:"deliveryType" valid:"basaldeliverytype"`
	ScheduleName string          `json:"scheduleName" bson:"scheduleName" valid:"omitempty,required"`
	Rate         float64         `json:"rate" bson:"rate" valid:"omitempty,basalrate"`
	Duration     int             `json:"duration" bson:"duration" valid:"omitempty,basalduration"`
	Insulin      string          `json:"insulin" bson:"insulin,omitempty" valid:"omitempty,basalinsulin"`
	Value        int             `json:"value" bson:"value,omitempty" valid:"omitempty,basalvalue"`
	Suppressed   *SupressedBasal `json:"suppressed" bson:"suppressed,omitempty" valid:"omitempty,required"`
	Base         `bson:",inline"`
}

//SupressedBasal represents a suppressed basal portion of a basal
type SupressedBasal struct {
	Type         string  `json:"type" bson:"type" valid:"required"`
	DeliveryType string  `json:"deliveryType" bson:"deliveryType" valid:"basaldeliverytype"`
	ScheduleName string  `json:"scheduleName" bson:"scheduleName" valid:"omitempty,required"`
	Rate         float64 `json:"rate" bson:"rate" valid:"omitempty,basalrate"`
}

const (
	//BasalName is the given name for the type of a `Basal` datum
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

	injectedDelivery, scheduledDelivery, suspendDelivery, tempDelivery = "injected", "scheduled", "suspend", "temp"

	levemirInsulin, lantusInsulin = "levemir", "lantus"

	rateValidationLowerLimit     = 0.0
	durationValidationLowerLimit = 0
	valueValidationLowerLimit    = 0
	validationGreaterThanInt     = "Must be greater than %d"
)

var (
	allowedDeliveryTypes = map[string]string{injectedDelivery: injectedDelivery, scheduledDelivery: scheduledDelivery, suspendDelivery: suspendDelivery, tempDelivery: tempDelivery}
	allowedInsulins      = map[string]string{levemirInsulin: levemirInsulin, lantusInsulin: lantusInsulin}

	basalFailureReasons = validate.ErrorReasons{
		deliveryTypeTag: fmt.Sprintf("Must be one of %s,%s,%s,%s", injectedDelivery, scheduledDelivery, suspendDelivery, tempDelivery),
		rateTag:         fmt.Sprintf("Must be greater than %f", rateValidationLowerLimit),
		durationTag:     fmt.Sprintf(validationGreaterThanInt, durationValidationLowerLimit),
		valueTag:        fmt.Sprintf(validationGreaterThanInt, valueValidationLowerLimit),
		insulinTag:      fmt.Sprintf("Must be one of %s,%s", levemirInsulin, lantusInsulin),
	}
)

//BuildBasal will build a Basal record
func BuildBasal(obj map[string]interface{}) (*Basal, *Error) {

	base, errs := BuildBase(obj)
	cast := NewCaster(errs)

	basal := &Basal{
		DeliveryType: cast.ToString(deliveryTypeField, obj[deliveryTypeField]),
		ScheduleName: cast.ToString(scheduleNameField, obj[scheduleNameField]),
		Base:         base,
	}

	if obj[rateField] != nil {
		basal.Rate = cast.ToFloat64(rateField, obj[rateField])
	}
	if obj[durationField] != nil {
		basal.Duration = cast.ToInt(durationField, obj[durationField])
	}
	if obj[insulinField] != nil {
		basal.Insulin = cast.ToString(insulinField, obj[insulinField])
	}

	if validationErrors := validator.Struct(basal); len(validationErrors) > 0 {
		errs.AppendError(validationErrors.GetError(basalFailureReasons))
	}

	if errs.IsEmpty() {
		return basal, nil
	}
	return basal, errs
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

//Selector will return the `unique` fields used in upserts
func (b *Basal) Selector() interface{} {

	unique := map[string]interface{}{}
	unique[deliveryTypeField] = b.DeliveryType
	unique[scheduleNameField] = b.ScheduleName
	unique[deviceTimeField] = b.Time
	unique[TypeField] = b.Type
	return unique
}
