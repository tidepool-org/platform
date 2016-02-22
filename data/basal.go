package data

//Basal represents a basal device data record
type Basal struct {
	DeliveryType string          `json:"deliveryType" bson:"deliveryType" valid:"required"`
	Rate         float64         `json:"rate" bson:"rate" valid:"required"`
	Duration     int             `json:"duration" bson:"duration" valid:"required"`
	Suppressed   *SupressedBasal `json:"suppressed" bson:"suppressed,omitempty"`
	Base         `bson:",inline"`
}

//SupressedBasal represents a suppressed basal portion of a basal
type SupressedBasal struct {
	Type         string  `json:"type" bson:"type" valid:"required"`
	DeliveryType string  `json:"deliveryType" bson:"deliveryType" valid:"required"`
	Rate         float64 `json:"rate" bson:"rate" valid:"required"`
}

//BuildBasal will build a Basal record
func BuildBasal(obj map[string]interface{}) (*Basal, *Error) {

	const (
		deliveryTypeField = "deliveryType"
		insulinField      = "insulin"
		rateField         = "rate"
		durationField     = "duration"
	)

	base, errs := BuildBase(obj)
	cast := NewCaster(errs)

	basal := &Basal{
		Rate:         cast.ToFloat64(rateField, obj[rateField]),
		Duration:     cast.ToInt(durationField, obj[durationField]),
		DeliveryType: cast.ToString(deliveryTypeField, obj[deliveryTypeField]),
		Base:         base,
	}

	_, err := validator.Validate(basal)
	errs.AppendError(err)
	if errs.IsEmpty() {
		return basal, nil
	}
	return basal, errs
}
