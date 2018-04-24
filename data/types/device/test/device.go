package test

import (
	"github.com/tidepool-org/platform/data/types/device"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
)

func NewDevice() *device.Device {
	datum := &device.Device{}
	datum.Base = *testDataTypes.NewBase()
	datum.Type = "deviceEvent"
	datum.SubType = testDataTypes.NewType()
	return datum
}

func CloneDevice(datum *device.Device) *device.Device {
	if datum == nil {
		return nil
	}
	clone := &device.Device{}
	clone.Base = *testDataTypes.CloneBase(&datum.Base)
	clone.SubType = datum.SubType
	return clone
}
