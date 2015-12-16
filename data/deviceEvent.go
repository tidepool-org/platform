package data

type DeviceEvent struct {
	SubType string `json:"subType" valid:"required"`
	Base
}

func BuildDeviceEvent(t map[string]interface{}) (*DeviceEvent, error) {

	const (
		subTypeField = "subType"
	)

	base, err := buildBase(t)
	if err != nil {
		return nil, err
	}

	deviceEvent := &DeviceEvent{
		SubType: t[subTypeField].(string),
		Base:    base,
	}

	valid, err := validator.Validate(deviceEvent)

	if valid {
		return deviceEvent, nil
	}
	return nil, err
}
