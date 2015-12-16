package data

type DeviceEvent struct {
	SubType string `json:"subType" valid:"required"`
	Base
}

func BuildDeviceEvent(obj map[string]interface{}) (*DeviceEvent, []error) {

	const (
		sub_type_field = "subType"
	)

	var errs buildErrors

	base := buildBase(obj, &errs)

	subType, ok := obj[sub_type_field].(string)
	if !ok {
		errs.addFeildError(sub_type_field, obj[sub_type_field])
	}

	deviceEvent := &DeviceEvent{
		SubType: subType,
		Base:    base,
	}

	_, err := validator.Validate(deviceEvent)
	errs.addError(err)
	return deviceEvent, errs
}

func (this *DeviceEvent) Validate() error {
	_, err := validator.Validate(this)
	return err
}
