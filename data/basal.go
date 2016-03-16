package data

import (
	"reflect"
	"strings"

	valid "github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
)

func init() {
	validator.RegisterValidation("basalrate", BasalRateValidator)
	validator.RegisterValidation("basalduration", BasalDurationValidator)
	validator.RegisterValidation("basaldeliverytype", BasalDeliveryTypeValidator)
	validator.RegisterValidation("basalinjection", BasalInjectionValidator)
	validator.RegisterValidation("basalinjectionvalue", BasalInjectionValueValidator)
}

//Basal represents a basal device data record
type Basal struct {
	DeliveryType string          `json:"deliveryType" bson:"deliveryType" valid:"basaldeliverytype"`
	ScheduleName string          `json:"scheduleName" bson:"scheduleName" valid:"omitempty,required"`
	Rate         float64         `json:"rate" bson:"rate" valid:"omitempty,basalrate"`
	Duration     int             `json:"duration" bson:"duration" valid:"omitempty,basalduration"`
	Insulin      string          `json:"insulin" bson:"insulin,omitempty" valid:"omitempty,basalinjection"`
	Value        int             `json:"value" bson:"value,omitempty" valid:"omitempty,basalinjectionvalue"`
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

	injectedDelivery, scheduledDelivery, suspendDelivery, tempDelivery = "injected", "scheduled", "suspend", "temp"

	levemirInsulin, lantusInsulin = "levemir", "lantus"
)

var (
	allowedDeliveryTypes = map[string]string{injectedDelivery: injectedDelivery, scheduledDelivery: scheduledDelivery, suspendDelivery: suspendDelivery, tempDelivery: tempDelivery}
	allowedInsulins      = map[string]string{levemirInsulin: levemirInsulin, lantusInsulin: lantusInsulin}
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

	errs.AppendError(validator.ValidateStruct(basal))
	if errs.IsEmpty() {
		return basal, nil
	}
	return basal, errs
}

func BasalRateValidator(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	if rate, ok := field.Interface().(float64); ok {
		if rate > 0 {
			return true
		}
	}
	return false
}

func BasalDurationValidator(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	if duration, ok := field.Interface().(int); ok {
		if duration > 0 {
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

func BasalInjectionValidator(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	if insulin, ok := field.Interface().(string); ok {
		if _, ok = allowedInsulins[strings.ToLower(insulin)]; ok {
			return true
		}
	}
	return false
}

func BasalInjectionValueValidator(v *valid.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	if value, ok := field.Interface().(int); ok {
		if value > 0 {
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
	unique[typeField] = b.Type
	return unique
}
