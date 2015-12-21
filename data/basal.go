package data

type Basal struct {
	DeliveryType string          `json:"deliveryType" valid:"required"`
	Rate         float64         `json:"rate" valid:"required"`
	Duration     int             `json:"duration" valid:"required"`
	Suppressed   *SupressedBasal `json:"suppressed"`
	Base
}

type SupressedBasal struct {
	Type         string  `json:"type" valid:"required"`
	DeliveryType string  `json:"deliveryType" valid:"required"`
	Rate         float64 `json:"rate" valid:"required"`
}

func BuildBasal(obj map[string]interface{}) (*Basal, *DataError) {

	const (
		delivery_type_field = "deliveryType"
		insulin_field       = "insulin"
		rate_field          = "rate"
		duration_field      = "duration"
	)

	base, errs := BuildBase(obj)
	cast := NewCaster(errs)

	basal := &Basal{
		Rate:         cast.ToFloat64(rate_field, obj[rate_field]),
		Duration:     cast.ToInt(duration_field, obj[duration_field]),
		DeliveryType: cast.ToString(delivery_type_field, obj[delivery_type_field]),
		Base:         base,
	}

	_, err := validator.Validate(basal)
	errs.AppendError(err)
	if errs.IsEmpty() {
		return basal, nil
	}
	return basal, errs
}
