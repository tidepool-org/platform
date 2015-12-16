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

func BuildBasal(t map[string]interface{}) (*Basal, error) {

	const (
		deliveryTypeField = "deliveryType"
		insulinField      = "insulin"
		valueField        = "value"
		durationField     = "duration"
	)

	base, err := buildBase(t)
	if err != nil {
		return nil, err
	}

	basal := &Basal{
		Insulin:      t[insulinField].(string),
		Value:        t[valueField].(float32),
		Duration:     t[durationField].(int64),
		DeliveryType: t[deliveryTypeField].(string),
		Base:         base,
	}

	valid, err := validator.Validate(basal)

	if valid {
		return basal, nil
	}
	return nil, err
}
