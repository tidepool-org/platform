package test

import (
	"math"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomEGVsResponse() *dexcom.EGVsResponse {
	datum := dexcom.NewEGVsResponse()
	datum.RateUnit = pointer.FromString(test.RandomStringFromArray(dexcom.EGVsResponseRateUnits()))
	datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.EGVsResponseUnits()))
	datum.EGVs = RandomEGVs(datum.Unit, 0, 3)
	return datum
}

func CloneEGVsResponse(datum *dexcom.EGVsResponse) *dexcom.EGVsResponse {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewEGVsResponse()
	clone.RateUnit = pointer.CloneString(datum.RateUnit)
	clone.Unit = pointer.CloneString(datum.Unit)
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
	datum := dexcom.NewEGV(unit)
	datum.SystemTime = RandomSystemTime()
	datum.DisplayTime = RandomDisplayTime()
	switch *datum.Unit {
	case dexcom.EGVUnitMgdL:
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.EGVValueMgdLMinimum, dexcom.EGVValueMgdLMaximum))
		datum.RealTimeValue = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.EGVValueMgdLMinimum, dexcom.EGVValueMgdLMaximum))
		datum.SmoothedValue = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.EGVValueMgdLMinimum, dexcom.EGVValueMgdLMaximum))
	}
	datum.Status = pointer.FromString(test.RandomStringFromArray(dexcom.EGVStatuses()))
	datum.Trend = pointer.FromString(test.RandomStringFromArray(dexcom.EGVTrends()))
	datum.TrendRate = pointer.FromFloat64(test.RandomFloat64())
	datum.TransmitterID = pointer.FromString(RandomTransmitterID())
	datum.TransmitterTicks = pointer.FromInt(test.RandomIntFromRange(dexcom.EGVTransmitterTickMinimum, math.MaxInt32))
	return datum
}

func CloneEGV(datum *dexcom.EGV) *dexcom.EGV {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewEGV(datum.Unit)
	clone.SystemTime = CloneTime(datum.SystemTime)
	clone.DisplayTime = CloneTime(datum.DisplayTime)
	clone.Unit = pointer.CloneString(datum.Unit)
	clone.Value = pointer.CloneFloat64(datum.Value)
	clone.RealTimeValue = pointer.CloneFloat64(datum.RealTimeValue)
	clone.SmoothedValue = pointer.CloneFloat64(datum.SmoothedValue)
	clone.Status = pointer.CloneString(datum.Status)
	clone.Trend = pointer.CloneString(datum.Trend)
	clone.TrendRate = pointer.CloneFloat64(datum.TrendRate)
	clone.TransmitterID = pointer.CloneString(datum.TransmitterID)
	clone.TransmitterTicks = pointer.CloneInt(datum.TransmitterTicks)
	return clone
}
