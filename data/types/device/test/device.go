package test

import (
	"github.com/tidepool-org/platform/data/types/device"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/test"
)

func RandomDevice() *device.Device {
	datum := randomDevice()
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "deviceEvent"
	return datum
}

func RandomDeviceForParser() *device.Device {
	datum := randomDevice()
	datum.Base = *dataTypesTest.RandomBaseForParser()
	datum.Type = "deviceEvent"
	return datum
}

func randomDevice() *device.Device {
	datum := &device.Device{}
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

func NewObjectFromDevice(datum *device.Device, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := dataTypesTest.NewObjectFromBase(&datum.Base, objectFormat)
	object["subType"] = test.NewObjectFromString(datum.SubType, objectFormat)
	return object
}
