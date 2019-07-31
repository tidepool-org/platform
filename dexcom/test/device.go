package test

import (
	"math"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomDevicesResponse() *dexcom.DevicesResponse {
	datum := dexcom.NewDevicesResponse()
	datum.Devices = RandomDevices(0, 3)
	return datum
}

func CloneDevicesResponse(datum *dexcom.DevicesResponse) *dexcom.DevicesResponse {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewDevicesResponse()
	clone.Devices = CloneDevices(datum.Devices)
	return clone
}

func RandomDevices(minimumLength int, maximumLength int) *dexcom.Devices {
	datum := make(dexcom.Devices, test.RandomIntFromRange(minimumLength, maximumLength))
	for index := range datum {
		datum[index] = RandomDevice()
	}
	return &datum
}

func CloneDevices(datum *dexcom.Devices) *dexcom.Devices {
	if datum == nil {
		return nil
	}
	clone := make(dexcom.Devices, len(*datum))
	for index, d := range *datum {
		clone[index] = CloneDevice(d)
	}
	return &clone
}

func RandomDevice() *dexcom.Device {
	datum := dexcom.NewDevice()
	datum.LastUploadDate = RandomTime()
	datum.AlertScheduleList = RandomAlertSchedules(1, 3)
	datum.UDI = pointer.FromString(RandomUDI())
	datum.SerialNumber = pointer.FromString(RandomSerialNumber())
	datum.TransmitterID = pointer.FromString(RandomTransmitterID())
	datum.TransmitterGeneration = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceTransmitterGenerations()))
	datum.DisplayDevice = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceDisplayDevices()))
	datum.SoftwareVersion = pointer.FromString(RandomSoftwareVersion())
	datum.SoftwareNumber = pointer.FromString(RandomSoftwareNumber())
	datum.Language = pointer.FromString(RandomLanguage())
	datum.IsMmolDisplayMode = pointer.FromBool(test.RandomBool())
	datum.IsBlindedMode = pointer.FromBool(test.RandomBool())
	datum.Is24HourMode = pointer.FromBool(test.RandomBool())
	datum.DisplayTimeOffset = pointer.FromInt(test.RandomIntFromRange(-math.MaxInt32, math.MaxInt32))
	datum.SystemTimeOffset = pointer.FromInt(test.RandomIntFromRange(-math.MaxInt32, math.MaxInt32))
	return datum
}

func CloneDevice(datum *dexcom.Device) *dexcom.Device {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewDevice()
	clone.LastUploadDate = CloneTime(datum.LastUploadDate)
	clone.AlertScheduleList = CloneAlertSchedules(datum.AlertScheduleList)
	clone.UDI = pointer.CloneString(datum.UDI)
	clone.SerialNumber = pointer.CloneString(datum.SerialNumber)
	clone.TransmitterID = pointer.CloneString(datum.TransmitterID)
	clone.TransmitterGeneration = pointer.CloneString(datum.TransmitterGeneration)
	clone.DisplayDevice = pointer.CloneString(datum.DisplayDevice)
	clone.SoftwareVersion = pointer.CloneString(datum.SoftwareVersion)
	clone.SoftwareNumber = pointer.CloneString(datum.SoftwareNumber)
	clone.Language = pointer.CloneString(datum.Language)
	clone.IsMmolDisplayMode = pointer.CloneBool(datum.IsMmolDisplayMode)
	clone.IsBlindedMode = pointer.CloneBool(datum.IsBlindedMode)
	clone.Is24HourMode = pointer.CloneBool(datum.Is24HourMode)
	clone.DisplayTimeOffset = pointer.CloneInt(datum.DisplayTimeOffset)
	clone.SystemTimeOffset = pointer.CloneInt(datum.SystemTimeOffset)
	return clone
}

func RandomUDI() string {
	return test.RandomString()
}

func RandomSerialNumber() string {
	return test.RandomString()
}

func RandomSoftwareVersion() string {
	return test.RandomString()
}

func RandomSoftwareNumber() string {
	return test.RandomString()
}

func RandomLanguage() string {
	return test.RandomString()
}
