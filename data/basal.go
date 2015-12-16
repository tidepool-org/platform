package data

type Basal struct {
	DeliveryType string          `json:"deliveryType" valid:"required"`
	Insulin      string          `json:"insulin" valid:"required"`
	Value        float32         `json:"value" valid:"required"`
	Duration     int64           `json:"duration" valid:"required"`
	Suppressed   *SupressedBasal `json:"suppressed"`
	Base
}

type SupressedBasal struct {
	Type         string  `json:"type" valid:"required"`
	DeliveryType string  `json:"deliveryType" valid:"required"`
	Value        float32 `json:"value" valid:"required"`
}

func BuildBasal(obj map[string]interface{}) (*Basal, *DataError) {

	const (
		delivery_type_field = "deliveryType"
		insulin_field       = "insulin"
		value_field         = "value"
		duration_field      = "duration"
	)

	base, errs := buildBase(obj)

	insulin, ok := obj[insulin_field].(string)
	if !ok {
		errs.AppendFieldError(insulin_field, obj[insulin_field])
	}

	value, ok := obj[value_field].(float32)
	if !ok {
		errs.AppendFieldError(value_field, obj[value_field])
	}

	duration, ok := obj[duration_field].(int64)
	if !ok {
		errs.AppendFieldError(duration_field, obj[duration_field])
	}

	deliveryType, ok := obj[delivery_type_field].(string)
	if !ok {
		errs.AppendFieldError(delivery_type_field, obj[delivery_type_field])
	}

	basal := &Basal{
		Insulin:      insulin,
		Value:        value,
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
