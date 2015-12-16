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

func BuildBasal(obj map[string]interface{}) (*Basal, error) {

	const (
		delivery_type_field = "deliveryType"
		insulin_field       = "insulin"
		value_field         = "value"
		duration_field      = "duration"
	)

	base, err := buildBase(obj)
	if err != nil {
		return nil, err
	}

	basal := &Basal{
		Insulin:      obj[insulin_field].(string),
		Value:        obj[value_field].(float32),
		Duration:     obj[duration_field].(int64),
		DeliveryType: obj[delivery_type_field].(string),
		Base:         base,
	}

	valid, err := validator.Validate(basal)

	if valid {
		return basal, nil
	}
	return nil, err
}
