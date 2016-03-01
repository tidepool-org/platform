package data

//Basal represents a basal device data record
type Basal struct {
	BaseBasal  `bson:",inline"`
	Suppressed *SupressedBasal `json:"suppressed" bson:"suppressed,omitempty"`
	Previous   *BaseBasal      `json:"previous" bson:"previous,omitempty"`
}

//BaseBasal represents the standard basal type fields
type BaseBasal struct {
	DeliveryType string  `json:"deliveryType" bson:"deliveryType" valid:"required"`
	ScheduleName string  `json:"scheduleName" bson:"scheduleName" valid:"required"`
	Rate         float64 `json:"rate" bson:"rate" valid:"required"`
	Duration     int     `json:"duration" bson:"duration" valid:"required"`
	Base         `bson:",inline"`
}

//SupressedBasal represents a suppressed basal portion of a basal
type SupressedBasal struct {
	Type         string  `json:"type" bson:"type" valid:"required"`
	DeliveryType string  `json:"deliveryType" bson:"deliveryType" valid:"required"`
	ScheduleName string  `json:"scheduleName" bson:"scheduleName" valid:"required"`
	Rate         float64 `json:"rate" bson:"rate" valid:"required"`
}

//BuildBasal will build a Basal record
func BuildBasal(obj map[string]interface{}) (*Basal, *Error) {

	const (
		deliveryTypeField = "deliveryType"
		scheduleNameField = "scheduleName"
		insulinField      = "insulin"
		rateField         = "rate"
		durationField     = "duration"
	)

	base, errs := BuildBase(obj)
	cast := NewCaster(errs)

	basal := &Basal{
		BaseBasal: BaseBasal{
			Rate:         cast.ToFloat64(rateField, obj[rateField]),
			Duration:     cast.ToInt(durationField, obj[durationField]),
			DeliveryType: cast.ToString(deliveryTypeField, obj[deliveryTypeField]),
			ScheduleName: cast.ToString(scheduleNameField, obj[scheduleNameField]),
			Base:         base,
		},
		Previous: &BaseBasal{
			Rate:         cast.ToFloat64(rateField, obj[rateField]),
			Duration:     cast.ToInt(durationField, obj[durationField]),
			DeliveryType: cast.ToString(deliveryTypeField, obj[deliveryTypeField]),
			ScheduleName: cast.ToString(scheduleNameField, obj[scheduleNameField]),
			Base:         base,
		},
	}

	_, err := validator.ValidateStruct(basal)
	errs.AppendError(err)
	if errs.IsEmpty() {
		return basal, nil
	}
	return basal, errs
}
