package test

import (
	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomDevicesResponse() *dexcom.DevicesResponse {
	datum := dexcom.NewDevicesResponse()
	datum.RecordType = pointer.FromString(dexcom.DevicesResponseRecordType)
	datum.RecordVersion = pointer.FromString(dexcom.DevicesResponseRecordVersion)
	datum.UserID = pointer.FromString(test.RandomString())
	datum.Records = RandomDevices(1, 3)
	return datum
}

func CloneDevicesResponse(datum *dexcom.DevicesResponse) *dexcom.DevicesResponse {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewDevicesResponse()
	clone.RecordType = pointer.CloneString(datum.RecordType)
	clone.RecordVersion = pointer.CloneString(datum.RecordVersion)
	clone.UserID = pointer.CloneString(datum.UserID)
	clone.Records = CloneDevices(datum.Records)
	return clone
}

func NewObjectFromDevicesResponse(datum *dexcom.DevicesResponse, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.RecordType != nil {
		object["recordType"] = test.NewObjectFromString(*datum.RecordType, objectFormat)
	}
	if datum.RecordVersion != nil {
		object["recordVersion"] = test.NewObjectFromString(*datum.RecordVersion, objectFormat)
	}
	if datum.UserID != nil {
		object["userId"] = test.NewObjectFromString(*datum.UserID, objectFormat)
	}
	if datum.Records != nil {
		object["records"] = NewArrayFromDevices(datum.Records, objectFormat)
	}
	return object
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
	for index, datum := range *datum {
		clone[index] = CloneDevice(datum)
	}
	return &clone
}

func NewArrayFromDevices(datumArray *dexcom.Devices, objectFormat test.ObjectFormat) []interface{} {
	if datumArray == nil {
		return nil
	}
	array := make([]interface{}, len(*datumArray))
	for index, datum := range *datumArray {
		array[index] = NewObjectFromDevice(datum, objectFormat)
	}
	return array
}

func RandomDevice() *dexcom.Device {
	datum := dexcom.NewDevice()
	datum.LastUploadDate = RandomTime()
	datum.AlertSchedules = RandomAlertSchedules(1, 3)
	datum.TransmitterID = pointer.FromString(RandomTransmitterID())
	datum.TransmitterGeneration = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceTransmitterGenerations()))
	datum.DisplayDevice = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceDisplayDevices()))
	datum.DisplayApp = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceDisplayApps()))
	return datum
}

func CloneDevice(datum *dexcom.Device) *dexcom.Device {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewDevice()
	clone.LastUploadDate = CloneTime(datum.LastUploadDate)
	clone.AlertSchedules = CloneAlertSchedules(datum.AlertSchedules)
	clone.TransmitterID = pointer.CloneString(datum.TransmitterID)
	clone.TransmitterGeneration = pointer.CloneString(datum.TransmitterGeneration)
	clone.DisplayDevice = pointer.CloneString(datum.DisplayDevice)
	clone.DisplayApp = pointer.CloneString(datum.DisplayApp)
	return clone
}

func NewObjectFromDevice(datum *dexcom.Device, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.LastUploadDate != nil {
		object["lastUploadDate"] = test.NewObjectFromString(datum.LastUploadDate.String(), objectFormat)
	}
	if datum.AlertSchedules != nil {
		object["alertSchedules"] = NewArrayFromAlertSchedules(datum.AlertSchedules, objectFormat)
	}
	if datum.TransmitterGeneration != nil {
		object["transmitterGeneration"] = test.NewObjectFromString(*datum.TransmitterGeneration, objectFormat)
	}
	if datum.TransmitterID != nil {
		object["transmitterId"] = test.NewObjectFromString(*datum.TransmitterID, objectFormat)
	}
	if datum.DisplayDevice != nil {
		object["displayDevice"] = test.NewObjectFromString(*datum.DisplayDevice, objectFormat)
	}
	if datum.DisplayApp != nil {
		object["displayApp"] = test.NewObjectFromString(*datum.DisplayApp, objectFormat)
	}
	return object
}
