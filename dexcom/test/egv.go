package test

import (
	"math"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomEGVsResponse() *dexcom.EGVsResponse {
	datum := dexcom.NewEGVsResponse()
	unit := pointer.FromString(test.RandomStringFromArray(dexcom.EGVsResponseUnits()))
	datum.EGVs = RandomEGVs(unit, 0, 3)
	return datum
}

func CloneEGVsResponse(datum *dexcom.EGVsResponse) *dexcom.EGVsResponse {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewEGVsResponse()
	clone.EGVs = CloneEGVs(datum.EGVs)
	return clone
}

func RandomEGVs(unit *string, minimumLength int, maximumLength int) *dexcom.EGVs {
	datum := make(dexcom.EGVs, test.RandomIntFromRange(minimumLength, maximumLength))
	for index := range datum {
		datum[index] = RandomEGV(unit)
	}
	return &datum
}

func CloneEGVs(datum *dexcom.EGVs) *dexcom.EGVs {
	if datum == nil {
		return nil
	}
	clone := make(dexcom.EGVs, len(*datum))
	for index, d := range *datum {
		clone[index] = CloneEGV(d)
	}
	return &clone
}

func RandomEGV(unit *string) *dexcom.EGV {
	datum := dexcom.NewEGV()
	datum.Unit = unit
	datum.ID = pointer.FromString(test.RandomString())
	datum.SystemTime = RandomSystemTime()
	datum.DisplayTime = RandomDisplayTime()
	switch *datum.Unit {
	case dexcom.EGVUnitMgdL:
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.EGVValueMgdLMinimum, dexcom.EGVValueMgdLMaximum))
		datum.RateUnit = pointer.FromString(dexcom.EGVRateUnitMgdLMinute)
	case dexcom.EGVUnitMmolL:
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.EGVValueMmolLMinimum, dexcom.EGVValueMmolLMaximum))
		datum.RateUnit = pointer.FromString(dexcom.EGVRateUnitMmolLMinute)
	case dexcom.EGVUnitUnknown:
		datum.Value = nil
		datum.RateUnit = pointer.FromString(dexcom.EGVRateUnitUnknown)
	}
	datum.Status = pointer.FromString(test.RandomStringFromArray(dexcom.EGVStatuses()))
	datum.Trend = pointer.FromString(test.RandomStringFromArray(dexcom.EGVTrends()))
	datum.TrendRate = pointer.FromFloat64(test.RandomFloat64())
	datum.TransmitterID = pointer.FromString(RandomTransmitterID())
	datum.TransmitterTicks = pointer.FromInt(test.RandomIntFromRange(dexcom.EGVTransmitterTickMinimum, math.MaxInt32))
	datum.TransmitterGeneration = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceTransmitterGenerations()))
	datum.DisplayDevice = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceDisplayDevices()))
	return datum
}

func CloneEGV(datum *dexcom.EGV) *dexcom.EGV {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewEGV()
	clone.ID = pointer.CloneString(datum.ID)
	clone.SystemTime = CloneTime(datum.SystemTime)
	clone.DisplayTime = CloneTime(datum.DisplayTime)
	clone.Unit = pointer.CloneString(datum.Unit)
	clone.RateUnit = pointer.CloneString(datum.RateUnit)
	clone.Value = pointer.CloneFloat64(datum.Value)
	clone.Status = pointer.CloneString(datum.Status)
	clone.Trend = pointer.CloneString(datum.Trend)
	clone.TrendRate = pointer.CloneFloat64(datum.TrendRate)
	clone.TransmitterID = pointer.CloneString(datum.TransmitterID)
	clone.TransmitterTicks = pointer.CloneInt(datum.TransmitterTicks)
	clone.TransmitterGeneration = pointer.CloneString(datum.TransmitterGeneration)
	clone.DisplayDevice = pointer.CloneString(datum.DisplayDevice)
	return clone
}
