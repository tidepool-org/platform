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
	datum.Records = RandomEvents(1, 3)
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

func NewObjectFromEventsResponse(datum *dexcom.EventsResponse, objectFormat test.ObjectFormat) map[string]interface{} {
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
		object["records"] = NewArrayFromEvents(datum.Records, objectFormat)
	}
	return object
}

func RandomEvents(minimumLength int, maximumLength int) *dexcom.Events {
	datum := make(dexcom.Events, test.RandomIntFromRange(minimumLength, maximumLength))
	for index := range datum {
		datum[index] = RandomEventWithType(test.RandomStringFromArray(dexcom.EventTypes()))
	}
	return &datum
}

func CloneEvents(datum *dexcom.Events) *dexcom.Events {
	if datum == nil {
		return nil
	}
	clone := make(dexcom.Events, len(*datum))
	for index, datum := range *datum {
		clone[index] = CloneEvent(datum)
	}
	return &clone
}

func NewArrayFromEvents(datumArray *dexcom.Events, objectFormat test.ObjectFormat) []interface{} {
	if datumArray == nil {
		return nil
	}
	array := make([]interface{}, len(*datumArray))
	for index, datum := range *datumArray {
		array[index] = NewObjectFromEvent(datum, objectFormat)
	}
	return array
}

func RandomEvent() *dexcom.Event {
	return RandomEventWithType(test.RandomStringFromArray(dexcom.EventTypes()))
}

func RandomEventWithType(tipe string) *dexcom.Event {
	datum := dexcom.NewEvent()
	datum.RecordID = pointer.FromString(test.RandomString())
	datum.SystemTime = RandomSystemTime()
	datum.DisplayTime = RandomDisplayTime()
	datum.EventStatus = pointer.FromString(test.RandomStringFromArray(dexcom.EventStatuses()))
	datum.EventType = pointer.FromString(tipe)
	switch tipe {
	case dexcom.EventTypeUnknown:
	case dexcom.EventTypeInsulin:
		datum.EventSubType = pointer.FromString(test.RandomStringFromArray(dexcom.EventSubTypesInsulin()))
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.EventUnitsInsulin()))
		switch *datum.Unit {
		case dexcom.EventUnitInsulinUnits:
			datum.Value = stringPointerFromRandomFloat64FromRange(dexcom.EventValueInsulinUnitsMinimum, dexcom.EventValueInsulinUnitsMaximum)
		}
	case dexcom.EventTypeCarbs:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.EventUnitsCarbs()))
		switch *datum.Unit {
		case dexcom.EventUnitCarbsGrams:
			datum.Value = stringPointerFromRandomFloat64FromRange(dexcom.EventValueCarbsGramsMinimum, dexcom.EventValueCarbsGramsMaximum)
		}
	case dexcom.EventTypeExercise:
		datum.EventSubType = pointer.FromString(test.RandomStringFromArray(dexcom.EventSubTypesExercise()))
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.EventUnitsExercise()))
		switch *datum.Unit {
		case dexcom.EventUnitExerciseMinutes:
			datum.Value = stringPointerFromRandomFloat64FromRange(dexcom.EventValueExerciseMinutesMinimum, dexcom.EventValueExerciseMinutesMaximum)
		}
	case dexcom.EventTypeHealth:
		datum.EventSubType = pointer.FromString(test.RandomStringFromArray(dexcom.EventSubTypesHealth()))
	case dexcom.EventTypeBloodGlucose:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.EventUnitsBloodGlucose()))
		switch *datum.Unit {
		case dexcom.EventUnitBloodGlucoseMgdL:
			datum.Value = stringPointerFromRandomFloat64FromRange(dexcom.EventValueBloodGlucoseMgdLMinimum, dexcom.EventValueBloodGlucoseMgdLMaximum)
		case dexcom.EventUnitBloodGlucoseMmolL:
			datum.Value = stringPointerFromRandomFloat64FromRange(dexcom.EventValueBloodGlucoseMmolLMinimum, dexcom.EventValueBloodGlucoseMmolLMaximum)
		}
	case dexcom.EventTypeNotes:
		datum.Value = pointer.FromString(test.RandomString())
	}
	datum.TransmitterGeneration = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceTransmitterGenerations()))
	datum.TransmitterID = pointer.FromString(RandomTransmitterID())
	datum.DisplayDevice = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceDisplayDevices()))
	return datum
}

func CloneEvent(datum *dexcom.Event) *dexcom.Event {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewEvent()
	clone.RecordID = pointer.CloneString(datum.RecordID)
	clone.SystemTime = CloneTime(datum.SystemTime)
	clone.DisplayTime = CloneTime(datum.DisplayTime)
	clone.EventStatus = pointer.CloneString(datum.EventStatus)
	clone.EventType = pointer.CloneString(datum.EventType)
	clone.EventSubType = pointer.CloneString(datum.EventSubType)
	clone.Unit = pointer.CloneString(datum.Unit)
	clone.Value = pointer.CloneString(datum.Value)
	clone.TransmitterGeneration = pointer.CloneString(datum.TransmitterGeneration)
	clone.TransmitterID = pointer.CloneString(datum.TransmitterID)
	clone.DisplayDevice = pointer.CloneString(datum.DisplayDevice)
	return clone
}

func NewObjectFromEvent(datum *dexcom.Event, objectFormat test.ObjectFormat) map[string]interface{} {
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
	if datum.EventStatus != nil {
		object["eventStatus"] = test.NewObjectFromString(*datum.EventStatus, objectFormat)
	}
	if datum.EventType != nil {
		object["eventType"] = test.NewObjectFromString(*datum.EventType, objectFormat)
	}
	if datum.EventSubType != nil {
		object["eventSubType"] = test.NewObjectFromString(*datum.EventSubType, objectFormat)
	}
	if datum.Unit != nil {
		object["unit"] = test.NewObjectFromString(*datum.Unit, objectFormat)
	}
	if datum.Value != nil {
		object["value"] = test.NewObjectFromString(*datum.Value, objectFormat)
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
	return object
}

func stringPointerFromRandomFloat64FromRange(minimum float64, maximum float64) *string {
	return StringPointerFromFloat64(test.RandomFloat64FromRange(minimum, maximum))
}

func StringPointerFromFloat64(value float64) *string {
	return pointer.FromString(fmt.Sprintf("%f", value))
}
