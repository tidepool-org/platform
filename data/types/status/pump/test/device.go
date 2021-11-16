package test

import (
	dataTypesStatusPump "github.com/tidepool-org/platform/data/types/status/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomDevice() *dataTypesStatusPump.Device {
	datum := dataTypesStatusPump.NewDevice()
	datum.ID = pointer.FromString(test.RandomStringFromRange(dataTypesStatusPump.DeviceIDLengthMinimum, dataTypesStatusPump.DeviceIDLengthMaximum))
	datum.Name = pointer.FromString(test.RandomStringFromRange(dataTypesStatusPump.DeviceNameLengthMinimum, dataTypesStatusPump.DeviceNameLengthMaximum))
	datum.Manufacturer = pointer.FromString(test.RandomStringFromRange(dataTypesStatusPump.DeviceManufacturerLengthMinimum, dataTypesStatusPump.DeviceManufacturerLengthMaximum))
	datum.Model = pointer.FromString(test.RandomStringFromRange(dataTypesStatusPump.DeviceModelLengthMinimum, dataTypesStatusPump.DeviceModelLengthMaximum))
	datum.FirmwareVersion = pointer.FromString(test.RandomStringFromRange(dataTypesStatusPump.DeviceVersionLengthMinimum, dataTypesStatusPump.DeviceVersionLengthMaximum))
	datum.HardwareVersion = pointer.FromString(test.RandomStringFromRange(dataTypesStatusPump.DeviceVersionLengthMinimum, dataTypesStatusPump.DeviceVersionLengthMaximum))
	datum.SoftwareVersion = pointer.FromString(test.RandomStringFromRange(dataTypesStatusPump.DeviceVersionLengthMinimum, dataTypesStatusPump.DeviceVersionLengthMaximum))
	return datum
}

func CloneDevice(datum *dataTypesStatusPump.Device) *dataTypesStatusPump.Device {
	if datum == nil {
		return nil
	}
	clone := dataTypesStatusPump.NewDevice()
	clone.ID = pointer.CloneString(datum.ID)
	clone.Name = pointer.CloneString(datum.Name)
	clone.Manufacturer = pointer.CloneString(datum.Manufacturer)
	clone.Model = pointer.CloneString(datum.Model)
	clone.FirmwareVersion = pointer.CloneString(datum.FirmwareVersion)
	clone.HardwareVersion = pointer.CloneString(datum.HardwareVersion)
	clone.SoftwareVersion = pointer.CloneString(datum.SoftwareVersion)
	return clone
}
