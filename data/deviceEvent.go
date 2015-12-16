package data

type DeviceEvent struct {
	SubType string `json:"subType" valid:"required"`
	Base
}

func BuildDeviceEvent(obj map[string]interface{}) (*DeviceEvent, *DataError) {

	const (
		sub_type_field = "subType"
	)

	base, errs := buildBase(obj)

	subType, ok := obj[sub_type_field].(string)
	if !ok {
		errs.AppendFieldError(sub_type_field, obj[sub_type_field])
	}

	deviceEvent := &DeviceEvent{
		SubType: subType,
		Base:    base,
	}

	_, err := validator.Validate(deviceEvent)
	errs.AppendError(err)
	if errs.IsEmpty() {
		return deviceEvent, nil
	}
	return deviceEvent, errs
}

func (this *DeviceEvent) Validate() error {
	_, err := validator.Validate(this)
	return err
}
