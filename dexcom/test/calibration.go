package test

import (
	"math"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomCalibrationsResponse() *dexcom.CalibrationsResponse {
	datum := dexcom.NewCalibrationsResponse()
	datum.RecordType = pointer.FromString(dexcom.CalibrationsResponseRecordType)
	datum.RecordVersion = pointer.FromString(dexcom.CalibrationsResponseRecordVersion)
	datum.UserID = pointer.FromString(test.RandomString())
	datum.Records = RandomCalibrationsWithUnit(test.RandomStringFromArray(dexcom.CalibrationUnits()), 1, 3)
	return datum
}

func CloneCalibrationsResponse(datum *dexcom.CalibrationsResponse) *dexcom.CalibrationsResponse {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewCalibrationsResponse()
	clone.RecordType = pointer.CloneString(datum.RecordType)
	clone.RecordVersion = pointer.CloneString(datum.RecordVersion)
	clone.UserID = pointer.CloneString(datum.UserID)
	clone.Records = CloneCalibrations(datum.Records)
	return clone
}

func NewObjectFromCalibrationsResponse(datum *dexcom.CalibrationsResponse, objectFormat test.ObjectFormat) map[string]interface{} {
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
		object["records"] = NewArrayFromCalibrations(datum.Records, objectFormat)
	}
	return object
}

func RandomCalibrationsWithUnit(unit string, minimumLength int, maximumLength int) *dexcom.Calibrations {
	datum := make(dexcom.Calibrations, test.RandomIntFromRange(minimumLength, maximumLength))
	for index := range datum {
		datum[index] = RandomCalibrationWithUnit(unit)
	}
	return &datum
}

func CloneCalibrations(datum *dexcom.Calibrations) *dexcom.Calibrations {
	if datum == nil {
		return nil
	}
	clone := make(dexcom.Calibrations, len(*datum))
	for index, datum := range *datum {
		clone[index] = CloneCalibration(datum)
	}
	return &clone
}

func NewArrayFromCalibrations(datumArray *dexcom.Calibrations, objectFormat test.ObjectFormat) []interface{} {
	if datumArray == nil {
		return nil
	}
	array := make([]interface{}, len(*datumArray))
	for index, datum := range *datumArray {
		array[index] = NewObjectFromCalibration(datum, objectFormat)
	}
	return array
}

func RandomCalibration() *dexcom.Calibration {
	return RandomCalibrationWithUnit(test.RandomStringFromArray(dexcom.CalibrationUnits()))
}

func RandomCalibrationWithUnit(unit string) *dexcom.Calibration {
	datum := dexcom.NewCalibration()
	datum.RecordID = pointer.FromString(test.RandomString())
	datum.SystemTime = RandomSystemTime()
	datum.DisplayTime = RandomDisplayTime()
	datum.Unit = pointer.FromString(unit)
	switch unit {
	case dexcom.CalibrationUnitMgdL:
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.CalibrationValueMgdLMinimum, dexcom.CalibrationValueMgdLMaximum))
	case dexcom.CalibrationUnitMmolL:
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.CalibrationValueMmolLMinimum, dexcom.CalibrationValueMmolLMaximum))
	case dexcom.CalibrationUnitUnknown:
		datum.Value = pointer.FromFloat64(test.RandomFloat64())
	}
	datum.TransmitterGeneration = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceTransmitterGenerations()))
	datum.TransmitterID = pointer.FromString(RandomTransmitterID())
	datum.TransmitterTicks = pointer.FromInt(test.RandomIntFromRange(dexcom.EGVTransmitterTickMinimum, math.MaxInt32))
	datum.DisplayDevice = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceDisplayDevices()))
	return datum
}

func CloneCalibration(datum *dexcom.Calibration) *dexcom.Calibration {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewCalibration()
	clone.RecordID = pointer.CloneString(datum.RecordID)
	clone.SystemTime = CloneTime(datum.SystemTime)
	clone.DisplayTime = CloneTime(datum.DisplayTime)
	clone.Unit = pointer.CloneString(datum.Unit)
	clone.Value = pointer.CloneFloat64(datum.Value)
	clone.TransmitterGeneration = pointer.CloneString(datum.TransmitterGeneration)
	clone.TransmitterID = pointer.CloneString(datum.TransmitterID)
	clone.TransmitterTicks = pointer.CloneInt(datum.TransmitterTicks)
	clone.DisplayDevice = pointer.CloneString(datum.DisplayDevice)
	return clone
}

func NewObjectFromCalibration(datum *dexcom.Calibration, objectFormat test.ObjectFormat) map[string]interface{} {
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
	if datum.Unit != nil {
		object["unit"] = test.NewObjectFromString(*datum.Unit, objectFormat)
	}
	if datum.Value != nil {
		object["value"] = test.NewObjectFromFloat64(*datum.Value, objectFormat)
	}
	if datum.TransmitterGeneration != nil {
		object["transmitterGeneration"] = test.NewObjectFromString(*datum.TransmitterGeneration, objectFormat)
	}
	if datum.TransmitterID != nil {
		object["transmitterId"] = test.NewObjectFromString(*datum.TransmitterID, objectFormat)
	}
	if datum.TransmitterTicks != nil {
		object["transmitterTicks"] = test.NewObjectFromInt(*datum.TransmitterTicks, objectFormat)
	}
	if datum.DisplayDevice != nil {
		object["displayDevice"] = test.NewObjectFromString(*datum.DisplayDevice, objectFormat)
	}
	return object
}
