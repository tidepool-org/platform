package test

import (
	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomAlertsResponse() *dexcom.AlertsResponse {
	datum := dexcom.NewAlertsResponse()
	datum.RecordType = pointer.FromString(dexcom.AlertsResponseRecordType)
	datum.RecordVersion = pointer.FromString(dexcom.AlertsResponseRecordVersion)
	datum.UserID = pointer.FromString(test.RandomString())
	datum.Records = RandomAlerts(1, 3)
	return datum
}

func CloneAlertsResponse(datum *dexcom.AlertsResponse) *dexcom.AlertsResponse {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewAlertsResponse()
	clone.RecordType = pointer.CloneString(datum.RecordType)
	clone.RecordVersion = pointer.CloneString(datum.RecordVersion)
	clone.UserID = pointer.CloneString(datum.UserID)
	clone.Records = CloneAlerts(datum.Records)
	return clone
}

func NewObjectFromAlertsResponse(datum *dexcom.AlertsResponse, objectFormat test.ObjectFormat) map[string]interface{} {
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
		object["records"] = NewArrayFromAlerts(datum.Records, objectFormat)
	}
	return object
}

func RandomAlerts(minimumLength int, maximumLength int) *dexcom.Alerts {
	datum := make(dexcom.Alerts, test.RandomIntFromRange(minimumLength, maximumLength))
	for index := range datum {
		datum[index] = RandomAlert()
	}
	return &datum
}

func CloneAlerts(datum *dexcom.Alerts) *dexcom.Alerts {
	if datum == nil {
		return nil
	}
	clone := make(dexcom.Alerts, len(*datum))
	for index, datum := range *datum {
		clone[index] = CloneAlert(datum)
	}
	return &clone
}

func NewArrayFromAlerts(datumArray *dexcom.Alerts, objectFormat test.ObjectFormat) []interface{} {
	if datumArray == nil {
		return nil
	}
	array := make([]interface{}, len(*datumArray))
	for index, datum := range *datumArray {
		array[index] = NewObjectFromAlert(datum, objectFormat)
	}
	return array
}

func RandomAlert() *dexcom.Alert {
	datum := dexcom.NewAlert()
	datum.RecordID = pointer.FromString(test.RandomString())
	datum.SystemTime = RandomSystemTime()
	datum.DisplayTime = RandomDisplayTime()
	datum.AlertName = pointer.FromString(test.RandomStringFromArray(dexcom.AlertNames()))
	datum.AlertState = pointer.FromString(test.RandomStringFromArray(dexcom.AlertStates()))
	datum.TransmitterGeneration = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceTransmitterGenerations()))
	datum.TransmitterID = pointer.FromString(RandomTransmitterID())
	datum.DisplayDevice = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceDisplayDevices()))
	datum.DisplayApp = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceDisplayApps()))
	return datum
}

func CloneAlert(datum *dexcom.Alert) *dexcom.Alert {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewAlert()
	clone.RecordID = pointer.CloneString(datum.RecordID)
	clone.SystemTime = CloneTime(datum.SystemTime)
	clone.DisplayTime = CloneTime(datum.DisplayTime)
	clone.AlertState = pointer.CloneString(datum.AlertState)
	clone.AlertName = pointer.CloneString(datum.AlertName)
	clone.TransmitterGeneration = pointer.CloneString(datum.TransmitterGeneration)
	clone.TransmitterID = pointer.CloneString(datum.TransmitterID)
	clone.DisplayDevice = pointer.CloneString(datum.DisplayDevice)
	clone.DisplayApp = pointer.CloneString(datum.DisplayApp)
	return clone
}

func NewObjectFromAlert(datum *dexcom.Alert, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.RecordID != nil {
		object["recordId"] = test.NewObjectFromString(*datum.RecordID, objectFormat)
	}
	if datum.SystemTime != nil {
		object["systemTime"] = test.NewObjectFromString(datum.SystemTime.String(), objectFormat)
	}
	if datum.DisplayTime != nil {
		object["displayTime"] = test.NewObjectFromString(datum.DisplayTime.String(), objectFormat)
	}
	if datum.AlertName != nil {
		object["alertName"] = test.NewObjectFromString(*datum.AlertName, objectFormat)
	}
	if datum.AlertState != nil {
		object["alertState"] = test.NewObjectFromString(*datum.AlertState, objectFormat)
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
