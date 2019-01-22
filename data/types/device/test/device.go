package test

import (
	"github.com/tidepool-org/platform/data/types/device"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
)

func NewDevice() *device.Device {
	datum := &device.Device{}
	datum.Base = *dataTypesTest.NewBase()
	datum.Type = "deviceEvent"
	datum.SubType = dataTypesTest.NewType()
	return datum
}

func CloneDevice(datum *device.Device) *device.Device {
	if datum == nil {
		return nil
	}
	clone := &device.Device{}
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.SubType = datum.SubType
	return clone
}
