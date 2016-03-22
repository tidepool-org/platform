package data

import "github.com/tidepool-org/platform/validate"

//DeviceEvent represents a deviceevent data record
type DeviceEvent struct {
	SubType string      `json:"subType" bson:"subType" valid:"required"`
	Status  string      `json:"status" bson:"status,omitempty" valid:"omitempty,required"`
	Reason  interface{} `json:"reason" bson:"reason,omitempty" valid:"-"`
	Base    `bson:",inline"`
}

const (
	//DeviceEventName is the given name for the type of a `DeviceEvent` datum
	DeviceEventName = "deviceEvent"

	statusField = "status"
	reasonField = "reason"
)

//BuildDeviceEvent will build a DeviceEvent record
func BuildDeviceEvent(obj map[string]interface{}) (*DeviceEvent, *Error) {

	base, errs := BuildBase(obj)
	cast := NewCaster(errs)

	deviceEvent := &DeviceEvent{
		SubType: cast.ToString(SubTypeField, obj[SubTypeField]),
		Status:  cast.ToString(statusField, obj[statusField]),
		Reason:  obj[reasonField],
		Base:    base,
	}

	if validationErrors := validator.Struct(deviceEvent); len(validationErrors) > 0 {
		errs.AppendError(validationErrors.GetError(validate.ErrorReasons{}))
	}

	if errs.IsEmpty() {
		return deviceEvent, nil
	}
	return deviceEvent, errs
}

//Selector will return the `unique` fields used in upserts
func (d *DeviceEvent) Selector() interface{} {

	unique := map[string]interface{}{}

	unique[SubTypeField] = d.SubType
	unique[deviceTimeField] = d.DeviceTime
	unique[TypeField] = d.Type
	return unique
}
