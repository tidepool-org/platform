package test

import (
	"fmt"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomEventsResponse() *dexcom.EventsResponse {
	datum := dexcom.NewEventsResponse()
	datum.RecordType = pointer.FromString(dexcom.EventsResponseRecordType)
	datum.RecordVersion = pointer.FromString(dexcom.EventsResponseRecordVersion)
	datum.UserID = pointer.FromString(test.RandomString())
	datum.Records = RandomEvents(0, 3)
	return datum
}

func CloneEventsResponse(datum *dexcom.EventsResponse) *dexcom.EventsResponse {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewEventsResponse()
	clone.RecordType = pointer.CloneString(datum.RecordType)
	clone.RecordVersion = pointer.CloneString(datum.RecordVersion)
	clone.UserID = pointer.CloneString(datum.UserID)
	clone.Records = CloneEvents(datum.Records)
	return clone
}

func RandomEvents(minimumLength int, maximumLength int) *dexcom.Events {
	datum := make(dexcom.Events, test.RandomIntFromRange(minimumLength, maximumLength))
	for index := range datum {
		datum[index] = RandomEvent(nil)
	}
	return &datum
}

func CloneEvents(datum *dexcom.Events) *dexcom.Events {
	if datum == nil {
		return nil
	}
	clone := make(dexcom.Events, len(*datum))
	for index, d := range *datum {
		clone[index] = CloneEvent(d)
	}
	return &clone
}

func RandomEvent(ofType *string) *dexcom.Event {
	datum := dexcom.NewEvent()
	datum.SystemTime = RandomSystemTime()
	datum.DisplayTime = RandomDisplayTime()
	if ofType != nil {
		datum.EventType = ofType
	} else {
		datum.EventType = pointer.FromString(test.RandomStringFromArray(dexcom.EventTypes()))
	}
	switch *datum.EventType {
	case dexcom.EventTypeCarbs:
		datum.Unit = pointer.FromString(dexcom.EventUnitCarbsGrams)
		datum.Value = pointer.FromString(
			fmt.Sprintf("%f", test.RandomFloat64FromRange(dexcom.EventValueCarbsGramsMinimum, dexcom.EventValueCarbsGramsMaximum)),
		)
	case dexcom.EventTypeExercise:
		datum.EventSubType = pointer.FromString(test.RandomStringFromArray(dexcom.EventSubTypesExercise()))
		datum.Unit = pointer.FromString(dexcom.EventUnitExerciseMinutes)
		datum.Value = pointer.FromString(
			fmt.Sprintf("%f", test.RandomFloat64FromRange(dexcom.EventValueExerciseMinutesMinimum, dexcom.EventValueExerciseMinutesMaximum)),
		)
	case dexcom.EventTypeHealth:
		datum.EventSubType = pointer.FromString(test.RandomStringFromArray(dexcom.EventSubTypesHealth()))
		datum.Value = pointer.FromString(test.RandomString())
	case dexcom.EventTypeInsulin:
		datum.EventSubType = pointer.FromString(test.RandomStringFromArray(dexcom.EventSubTypesInsulin()))
		datum.Unit = pointer.FromString(dexcom.EventUnitInsulinUnits)
		datum.Value = pointer.FromString(
			fmt.Sprintf("%f", test.RandomFloat64FromRange(dexcom.EventValueInsulinUnitsMinimum, dexcom.EventValueInsulinUnitsMaximum)),
		)
	case dexcom.EventTypeBG:
		datum.Unit = pointer.FromString(dexcom.EventUnitBGMgdL)
		datum.Value = pointer.FromString(
			fmt.Sprintf("%f", test.RandomFloat64FromRange(dexcom.EventValueMgdLMinimum, dexcom.EventValueMgdLMaximum)),
		)
	case dexcom.EventTypeNotes:
		datum.Unit = nil
		datum.Value = pointer.FromString(test.RandomString())
	case dexcom.EventTypeUnknown:
		datum.Unit = nil
		datum.Value = pointer.FromString(test.RandomString())
	}
	datum.RecordID = pointer.FromString(RandomEventID())
	datum.EventStatus = pointer.FromString(test.RandomStringFromArray(dexcom.EventStatuses()))
	datum.DisplayDevice = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceDisplayDevices()))
	datum.TransmitterGeneration = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceTransmitterGenerations()))
	datum.TransmitterID = pointer.FromString(RandomTransmitterID())
	return datum
}

func CloneEvent(datum *dexcom.Event) *dexcom.Event {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewEvent()
	clone.SystemTime = CloneTime(datum.SystemTime)
	clone.DisplayTime = CloneTime(datum.DisplayTime)
	clone.EventType = pointer.CloneString(datum.EventType)
	clone.EventSubType = pointer.CloneString(datum.EventSubType)
	clone.Unit = pointer.CloneString(datum.Unit)
	clone.Value = pointer.CloneString(datum.Value)
	clone.RecordID = pointer.CloneString(datum.RecordID)
	clone.EventStatus = pointer.CloneString(datum.EventStatus)
	return clone
}

func RandomEventID() string {
	return test.RandomString()
}
