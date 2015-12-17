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

	base, errs := buildBase(obj)

	rate, ok := obj[rate_field].(float64)
	if !ok {
		errs.AppendFieldError(rate_field, obj[rate_field])
	}

	duration, ok := obj[duration_field].(int)
	if !ok {

		duration_float64, ok := obj[duration_field].(float64)
		duration = int(duration_float64)
		if !ok {
			errs.AppendFieldError(duration_field, obj[duration_field])
		}
	}

	deliveryType, ok := obj[delivery_type_field].(string)
	if !ok {
		errs.AppendFieldError(delivery_type_field, obj[delivery_type_field])
	}

	basal := &Basal{
		Rate:         rate,
		Duration:     duration,
		DeliveryType: deliveryType,
		Base:         base,
	}

	_, err := validator.Validate(basal)
	errs.AppendError(err)
	if errs.IsEmpty() {
		return basal, nil
	}
	return basal, errs
}

func (this *Basal) Validate() error {
	_, err := validator.Validate(this)
	return err
}
