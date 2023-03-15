package test

import (
	"math"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomCalibrationsResponse() *dexcom.CalibrationsResponse {
	datum := dexcom.NewCalibrationsResponse()
	datum.Calibrations = RandomCalibrations(0, 3)
	return datum
}

func CloneCalibrationsResponse(datum *dexcom.CalibrationsResponse) *dexcom.CalibrationsResponse {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewCalibrationsResponse()
	clone.Calibrations = CloneCalibrations(datum.Calibrations)
	return clone
}

func RandomCalibrations(minimumLength int, maximumLength int) *dexcom.Calibrations {
	datum := make(dexcom.Calibrations, test.RandomIntFromRange(minimumLength, maximumLength))
	for index := range datum {
		datum[index] = RandomCalibration()
	}
	return &datum
}

func CloneCalibrations(datum *dexcom.Calibrations) *dexcom.Calibrations {
	if datum == nil {
		return nil
	}
	clone := make(dexcom.Calibrations, len(*datum))
	for index, d := range *datum {
		clone[index] = CloneCalibration(d)
	}
	return &clone
}

func RandomCalibration() *dexcom.Calibration {
	datum := dexcom.NewCalibration()
	datum.ID = pointer.FromString(test.RandomString())
	datum.SystemTime = RandomSystemTime()
	datum.DisplayTime = RandomDisplayTime()
	datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.CalibrationUnits()))
	switch *datum.Unit {
	case dexcom.CalibrationUnitMgdL:
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.CalibrationValueMgdLMinimum, dexcom.CalibrationValueMgdLMaximum))
	case dexcom.CalibrationUnitMmolL:
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.CalibrationValueMmolLMinimum, dexcom.CalibrationValueMmolLMaximum))
	}
	datum.TransmitterID = pointer.FromString(RandomTransmitterID())
	datum.TransmitterTicks = pointer.FromInt(test.RandomIntFromRange(dexcom.EGVTransmitterTickMinimum, math.MaxInt32))
	datum.TransmitterGeneration = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceTransmitterGenerations()))
	datum.DisplayDevice = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceDisplayDevices()))
	return datum
}

func CloneCalibration(datum *dexcom.Calibration) *dexcom.Calibration {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewCalibration()

	clone.ID = pointer.CloneString(datum.ID)
	clone.SystemTime = CloneTime(datum.SystemTime)
	clone.DisplayTime = CloneTime(datum.DisplayTime)
	clone.Unit = pointer.CloneString(datum.Unit)
	clone.Value = pointer.CloneFloat64(datum.Value)
	clone.TransmitterID = pointer.CloneString(datum.TransmitterID)
	clone.TransmitterTicks = pointer.CloneInt(datum.TransmitterTicks)
	clone.TransmitterGeneration = pointer.CloneString(datum.TransmitterGeneration)
	clone.DisplayDevice = pointer.CloneString(datum.DisplayDevice)
	return clone
}
