package test

import (
	dataTypesDevice "github.com/tidepool-org/platform/data/types/device"
	dataTypesDeviceTest "github.com/tidepool-org/platform/data/types/device/test"
	dataTypesDeviceTimechange "github.com/tidepool-org/platform/data/types/device/timechange"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &dataTypesDevice.Meta{
		Type:    "deviceEvent",
		SubType: "timeChange",
	}
}

func RandomTimeChange(deprecated bool) *dataTypesDeviceTimechange.TimeChange {
	datum := dataTypesDeviceTimechange.New()
	datum.Device = *dataTypesDeviceTest.RandomDevice()
	datum.SubType = "timeChange"
	if !deprecated {
		datum.From = RandomInfo()
		datum.Method = pointer.FromString(test.RandomStringFromArray(dataTypesDeviceTimechange.Methods()))
		datum.To = RandomInfo()
	} else {
		datum.Change = RandomChange()
	}
	return datum
}

func CloneTimeChange(datum *dataTypesDeviceTimechange.TimeChange) *dataTypesDeviceTimechange.TimeChange {
	if datum == nil {
		return nil
	}
	clone := dataTypesDeviceTimechange.New()
	clone.Device = *dataTypesDeviceTest.CloneDevice(&datum.Device)
	clone.From = CloneInfo(datum.From)
	clone.Method = pointer.CloneString(datum.Method)
	clone.To = CloneInfo(datum.To)
	clone.Change = CloneChange(datum.Change)
	return clone
}
