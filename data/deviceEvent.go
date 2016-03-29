package data

import "github.com/tidepool-org/platform/validate"

type DeviceEvent struct {
	SubType string      `json:"subType" bson:"subType" valid:"required"`
	Status  string      `json:"status" bson:"status,omitempty" valid:"omitempty,required"`
	Reason  interface{} `json:"reason" bson:"reason,omitempty" valid:"-"`
	Base    `bson:",inline"`
}

const (
	DeviceEventName = "deviceEvent"

	statusField = "status"
	reasonField = "reason"
)

func BuildDeviceEvent(datum Datum) (*DeviceEvent, *Error) {

	base, errs := BuildBase(datum)
	cast := NewCaster(errs)

	deviceEvent := &DeviceEvent{
		SubType: cast.ToString(SubTypeField, datum[SubTypeField]),
		Status:  cast.ToString(statusField, datum[statusField]),
		Reason:  datum[reasonField],
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
