package data

type DeviceEvent struct {
	SubType string `json:"subType" valid:"required"`
	Base
}

func BuildDeviceEvent(obj map[string]interface{}) (*DeviceEvent, error) {

	const (
		sub_type_field = "subType"
	)

	base, err := buildBase(obj)
	if err != nil {
		return nil, err
	}

	deviceEvent := &DeviceEvent{
		SubType: obj[sub_type_field].(string),
		Base:    base,
	}

	valid, err := validator.Validate(deviceEvent)

	if valid {
		return deviceEvent, nil
	}
	return nil, err
}
