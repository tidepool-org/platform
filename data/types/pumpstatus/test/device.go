package test

import (
	dataTypesPumpStatus "github.com/tidepool-org/platform/data/types/pumpstatus"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomDevice() *dataTypesPumpStatus.Device {
	datum := dataTypesPumpStatus.NewDevice()
	datum.ID = pointer.FromString(test.RandomStringFromRange(dataTypesPumpStatus.DeviceIDLengthMinimum, dataTypesPumpStatus.DeviceIDLengthMaximum))
	datum.Name = pointer.FromString(test.RandomStringFromRange(dataTypesPumpStatus.DeviceNameLengthMinimum, dataTypesPumpStatus.DeviceNameLengthMaximum))
	datum.Manufacturer = pointer.FromString(test.RandomStringFromRange(dataTypesPumpStatus.DeviceManufacturerLengthMinimum, dataTypesPumpStatus.DeviceManufacturerLengthMaximum))
	datum.Model = pointer.FromString(test.RandomStringFromRange(dataTypesPumpStatus.DeviceModelLengthMinimum, dataTypesPumpStatus.DeviceModelLengthMaximum))
	datum.FirmwareVersion = pointer.FromString(test.RandomStringFromRange(dataTypesPumpStatus.DeviceVersionLengthMinimum, dataTypesPumpStatus.DeviceVersionLengthMaximum))
	datum.HardwareVersion = pointer.FromString(test.RandomStringFromRange(dataTypesPumpStatus.DeviceVersionLengthMinimum, dataTypesPumpStatus.DeviceVersionLengthMaximum))
	datum.SoftwareVersion = pointer.FromString(test.RandomStringFromRange(dataTypesPumpStatus.DeviceVersionLengthMinimum, dataTypesPumpStatus.DeviceVersionLengthMaximum))
	return datum
}

func CloneDevice(datum *dataTypesPumpStatus.Device) *dataTypesPumpStatus.Device {
	if datum == nil {
		return nil
	}
	clone := dataTypesPumpStatus.NewDevice()
	clone.ID = pointer.CloneString(datum.ID)
	clone.Name = pointer.CloneString(datum.Name)
	clone.Manufacturer = pointer.CloneString(datum.Manufacturer)
	clone.Model = pointer.CloneString(datum.Model)
	clone.FirmwareVersion = pointer.CloneString(datum.FirmwareVersion)
	clone.HardwareVersion = pointer.CloneString(datum.HardwareVersion)
	clone.SoftwareVersion = pointer.CloneString(datum.SoftwareVersion)
	return clone
}
