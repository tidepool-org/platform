package data

import "github.com/tidepool-org/platform/validate"

type DeviceEvent struct {
	SubType string      `json:"subType" bson:"subType" valid:"required"`
	Status  string      `json:"status,omitempty" bson:"status,omitempty" valid:"omitempty,required"`
	Reason  interface{} `json:"reason,omitempty" bson:"reason,omitempty" valid:"-"`
	Base    `bson:",inline"`
}

const (
	DeviceEventName = "deviceEvent"

	subTypeField = "subType"
	statusField  = "status"
	reasonField  = "reason"
)

func BuildDeviceEvent(datum Datum, errs *DatumErrors) *DeviceEvent {

	base := BuildBase(datum, errs)

	deviceEvent := &DeviceEvent{
		SubType: cast.ToString(subTypeField, datum[subTypeField]),
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
