package test

import (
	"math"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomEGVsResponse() *dexcom.EGVsResponse {
	datum := dexcom.NewEGVsResponse()
	datum.RecordType = pointer.FromString(dexcom.EGVsResponseRecordType)
	datum.RecordVersion = pointer.FromString(dexcom.EGVsResponseRecordVersion)
	datum.UserID = pointer.FromString(test.RandomString())
	datum.Records = RandomEGVsWithUnit(test.RandomStringFromArray(dexcom.EGVUnits()), 1, 3)
	return datum
}

func CloneEGVsResponse(datum *dexcom.EGVsResponse) *dexcom.EGVsResponse {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewEGVsResponse()
	clone.RecordType = pointer.CloneString(datum.RecordType)
	clone.RecordVersion = pointer.CloneString(datum.RecordVersion)
	clone.UserID = pointer.CloneString(datum.UserID)
	clone.Records = CloneEGVs(datum.Records)
	return clone
}

func NewObjectFromEGVsResponse(datum *dexcom.EGVsResponse, objectFormat test.ObjectFormat) map[string]interface{} {
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
		object["records"] = NewArrayFromEGVs(datum.Records, objectFormat)
	}
	return object
}

func RandomEGVsWithUnit(unit string, minimumLength int, maximumLength int) *dexcom.EGVs {
	datum := make(dexcom.EGVs, test.RandomIntFromRange(minimumLength, maximumLength))
	for index := range datum {
		datum[index] = RandomEGVWithUnit(unit)
	}
	return &datum
}

func CloneEGVs(datum *dexcom.EGVs) *dexcom.EGVs {
	if datum == nil {
		return nil
	}
	clone := make(dexcom.EGVs, len(*datum))
	for index, datum := range *datum {
		clone[index] = CloneEGV(datum)
	}
	return &clone
}

func NewArrayFromEGVs(datumArray *dexcom.EGVs, objectFormat test.ObjectFormat) []interface{} {
	if datumArray == nil {
		return nil
	}
	array := make([]interface{}, len(*datumArray))
	for index, datum := range *datumArray {
		array[index] = NewObjectFromEGV(datum, objectFormat)
	}
	return array
}

func RandomEGV() *dexcom.EGV {
	return RandomEGVWithUnit(test.RandomStringFromArray(dexcom.EGVUnits()))
}

func RandomEGVWithUnit(unit string) *dexcom.EGV {
	datum := dexcom.NewEGV()
	datum.RecordID = pointer.FromString(test.RandomString())
	datum.SystemTime = RandomSystemTime()
	datum.DisplayTime = RandomDisplayTime()
	datum.Unit = pointer.FromString(unit)
	switch unit {
	case dexcom.EGVUnitMgdL:
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.EGVValueMgdLMinimum, dexcom.EGVValueMgdLMaximum))
		datum.RateUnit = pointer.FromString(dexcom.EGVRateUnitMgdLMinute)
	case dexcom.EGVUnitMmolL:
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.EGVValueMmolLMinimum, dexcom.EGVValueMmolLMaximum))
		datum.RateUnit = pointer.FromString(dexcom.EGVRateUnitMmolLMinute)
	case dexcom.EGVUnitUnknown:
		datum.Value = pointer.FromFloat64(test.RandomFloat64())
		datum.RateUnit = pointer.FromString(dexcom.EGVRateUnitUnknown)
	}
	datum.TrendRate = pointer.FromFloat64(test.RandomFloat64())
	datum.Status = pointer.FromString(test.RandomStringFromArray(dexcom.EGVStatuses()))
	datum.Trend = pointer.FromString(test.RandomStringFromArray(dexcom.EGVTrends()))
	datum.TransmitterGeneration = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceTransmitterGenerations()))
	datum.TransmitterID = pointer.FromString(RandomTransmitterID())
	datum.TransmitterTicks = pointer.FromInt(test.RandomIntFromRange(dexcom.EGVTransmitterTickMinimum, math.MaxInt32))
	datum.DisplayDevice = pointer.FromString(test.RandomStringFromArray(dexcom.DeviceDisplayDevices()))
	return datum
}

func CloneEGV(datum *dexcom.EGV) *dexcom.EGV {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewEGV()
	clone.RecordID = pointer.CloneString(datum.RecordID)
	clone.SystemTime = CloneTime(datum.SystemTime)
	clone.DisplayTime = CloneTime(datum.DisplayTime)
	clone.Unit = pointer.CloneString(datum.Unit)
	clone.Value = pointer.CloneFloat64(datum.Value)
	clone.RateUnit = pointer.CloneString(datum.RateUnit)
	clone.TrendRate = pointer.CloneFloat64(datum.TrendRate)
	clone.Status = pointer.CloneString(datum.Status)
	clone.Trend = pointer.CloneString(datum.Trend)
	clone.TransmitterGeneration = pointer.CloneString(datum.TransmitterGeneration)
	clone.TransmitterID = pointer.CloneString(datum.TransmitterID)
	clone.TransmitterTicks = pointer.CloneInt(datum.TransmitterTicks)
	clone.DisplayDevice = pointer.CloneString(datum.DisplayDevice)
	return clone
}

func NewObjectFromEGV(datum *dexcom.EGV, objectFormat test.ObjectFormat) map[string]interface{} {
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
	if datum.RateUnit != nil {
		object["rateUnit"] = test.NewObjectFromString(*datum.RateUnit, objectFormat)
	}
	if datum.TrendRate != nil {
		object["trendRate"] = test.NewObjectFromFloat64(*datum.TrendRate, objectFormat)
	}
	if datum.Status != nil {
		object["status"] = test.NewObjectFromString(*datum.Status, objectFormat)
	}
	if datum.Trend != nil {
		object["trend"] = test.NewObjectFromString(*datum.Trend, objectFormat)
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
