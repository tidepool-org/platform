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

const (
	DeviceEventName = "deviceEvent"

	statusField = "status"
	reasonField = "reason"
)

func BuildDeviceEvent(datum Datum, errs validate.ErrorProcessing) *DeviceEvent {

	deviceEvent := &DeviceEvent{
		Reason:  ToObject(reasonField, datum[reasonField], errs),
		SubType: ToString(SubTypeField, datum[SubTypeField], errs),
		Status:  ToString(statusField, datum[statusField], errs),
		Base:    BuildBase(datum, errs),
	}

	getPlatformValidator().Struct(deviceEvent, errs)

	return deviceEvent
}
