package data

//DeviceEvent represents a deviceevent data record
type DeviceEvent struct {
	SubType string `json:"subType" bson:"subType" valid:"required"`
	Base
}

//BuildDeviceEvent will build a DeviceEvent record
func BuildDeviceEvent(obj map[string]interface{}) (*DeviceEvent, *Error) {

	const (
		subTypeField = "subType"
	)

	base, errs := BuildBase(obj)
	cast := NewCaster(errs)

	deviceEvent := &DeviceEvent{
		SubType: cast.ToString(subTypeField, obj[subTypeField]),
		Base:    base,
	}

	_, err := validator.Validate(deviceEvent)
	errs.AppendError(err)
	if errs.IsEmpty() {
		return deviceEvent, nil
	}
	return deviceEvent, errs
}
