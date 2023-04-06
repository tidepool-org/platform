package test

import (
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
	datum.LastUploadDate = RandomTimeUTC()
	datum.AlertScheduleList = RandomAlertSchedules(1, 3)
	datum.TransmitterID = pointer.FromString(RandomTransmitterID())
	datum.TransmitterGeneration = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceTransmitterGenerations()))
	datum.DisplayDevice = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceDisplayDevices()))
	return datum
}

func CloneDevice(datum *dexcom.Device) *dexcom.Device {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewDevice()
	clone.LastUploadDate = CloneTime(datum.LastUploadDate)
	clone.AlertScheduleList = CloneAlertSchedules(datum.AlertScheduleList)
	clone.TransmitterID = pointer.CloneString(datum.TransmitterID)
	clone.TransmitterGeneration = pointer.CloneString(datum.TransmitterGeneration)
	clone.DisplayDevice = pointer.CloneString(datum.DisplayDevice)
	return clone
}
