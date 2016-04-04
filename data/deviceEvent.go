package data

import (
	"github.com/tidepool-org/platform/validate"
)

type DeviceEvent struct {
	SubType *string      `json:"subType" bson:"subType" valid:"required"`
	Status  *string      `json:"status,omitempty" bson:"status,omitempty" valid:"omitempty,required"`
	Reason  *interface{} `json:"reason,omitempty" bson:"reason,omitempty" valid:"-"`
	Base    `bson:",inline"`
}

const DeviceEventName = "deviceEvent"

var (
	deviceEventStatusField  = DatumField{Name: "status"}
	deviceEventReasonField  = DatumField{Name: "reason"}
	deviceEventSubTypeField = DatumField{Name: "subType"}
)

func BuildDeviceEvent(datum Datum, errs validate.ErrorProcessing) *DeviceEvent {

	deviceEvent := &DeviceEvent{
		Reason:  datum.ToObject(deviceEventReasonField.Name, errs),
		SubType: datum.ToString(deviceEventSubTypeField.Name, errs),
		Status:  datum.ToString(deviceEventStatusField.Name, errs),
		Base:    BuildBase(datum, errs),
	}

	getPlatformValidator().Struct(deviceEvent, errs)

	return deviceEvent
}
