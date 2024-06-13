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
	datum.Records = RandomAlerts(0, 5)
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
	for index, d := range *datum {
		clone[index] = CloneAlert(d)
	}
	return &clone
}

func RandomAlert() *dexcom.Alert {
	datum := dexcom.NewAlert()
	datum.RecordID = pointer.FromString(test.RandomString())
	datum.SystemTime = RandomSystemTime()
	datum.DisplayTime = RandomDisplayTime()
	datum.AlertName = pointer.FromString(test.RandomStringFromArray(dexcom.AlertNames()))
	datum.AlertState = pointer.FromString(test.RandomStringFromArray(dexcom.AlertStates()))
	datum.TransmitterID = pointer.FromString(RandomTransmitterID())
	datum.TransmitterGeneration = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceTransmitterGenerations()))
	datum.DisplayDevice = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceDisplayDevices()))
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
	clone.TransmitterID = pointer.CloneString(datum.TransmitterID)
	clone.TransmitterGeneration = pointer.CloneString(datum.TransmitterGeneration)
	clone.DisplayDevice = pointer.CloneString(datum.DisplayDevice)
	return clone
}
