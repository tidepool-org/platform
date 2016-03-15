package data

import (
	"reflect"
	"strings"

	valid "github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
)

func init() {
	validator.RegisterStructValidation(BasalValidation, Basal{})
}

//Basal represents a basal device data record
type Basal struct {
	DeliveryType string          `json:"deliveryType" bson:"deliveryType"`
	ScheduleName string          `json:"scheduleName" bson:"scheduleName" valid:"omitempty,required"`
	Rate         float64         `json:"rate" bson:"rate" valid:"omitempty,gte=0"`
	Duration     int             `json:"duration" bson:"duration" valid:"omitempty,gte=0"`
	Insulin      string          `json:"insulin" bson:"insulin,omitempty"`
	Value        int             `json:"value" bson:"value,omitempty"`
	Suppressed   *SupressedBasal `json:"suppressed" bson:"suppressed,omitempty" valid:"omitempty,required"`
	Base         `bson:",inline"`
}

//SupressedBasal represents a suppressed basal portion of a basal
type SupressedBasal struct {
	Type         string  `json:"type" bson:"type" valid:"required"`
	DeliveryType string  `json:"deliveryType" bson:"deliveryType" valid:"required"`
	ScheduleName string  `json:"scheduleName" bson:"scheduleName" valid:"omitempty,required"`
	Rate         float64 `json:"rate" bson:"rate" valid:"omitempty,gte=0"`
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

//BasalValidation used to validate the built basal record
// TODO: fixup signature so not using `validator.v8`
func BasalValidation(v *valid.Validate, structLevel *valid.StructLevel) {

	basal := structLevel.CurrentStruct.Interface().(Basal)

	if val, ok := allowedDeliveryTypes[strings.ToLower(basal.DeliveryType)]; !ok {
		structLevel.ReportError(reflect.ValueOf(basal.DeliveryType), "DeliveryType", deliveryTypeField, "deliverytypes")
	} else {
		switch val {
		case injectedDelivery:
			if _, ok := allowedInsulins[strings.ToLower(basal.Insulin)]; !ok {
				structLevel.ReportError(reflect.ValueOf(basal.Insulin), "Insulin", insulinField, "insulintypes")
			}
			if basal.Value <= 0 {
				structLevel.ReportError(reflect.ValueOf(basal.Value), "Value", valueField, "insulinvalue")
			}
		}
	}
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
