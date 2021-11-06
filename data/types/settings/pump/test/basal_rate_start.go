package test

import (
	dataTypesBasalTest "github.com/tidepool-org/platform/data/types/basal/test"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewBasalRateStart(startMinimum int) *pump.BasalRateStart {
	datum := pump.NewBasalRateStart()
	datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(pump.BasalRateStartRateMinimum, pump.BasalRateStartRateMaximum))
	if startMinimum == pump.BasalRateStartStartMinimum {
		datum.Start = pointer.FromInt(pump.BasalRateStartStartMinimum)
	} else {
		datum.Start = pointer.FromInt(test.RandomIntFromRange(startMinimum, pump.BasalRateStartStartMaximum))
	}
	return datum
}

func CloneBasalRateStart(datum *pump.BasalRateStart) *pump.BasalRateStart {
	if datum == nil {
		return nil
	}
	clone := pump.NewBasalRateStart()
	clone.Rate = pointer.CloneFloat64(datum.Rate)
	clone.Start = pointer.CloneInt(datum.Start)
	return clone
}

func NewBasalRateStartArray() *pump.BasalRateStartArray {
	datum := pump.NewBasalRateStartArray()
	*datum = append(*datum, NewBasalRateStart(pump.BasalRateStartStartMinimum))
	*datum = append(*datum, NewBasalRateStart(*datum.Last().Start+1))
	*datum = append(*datum, NewBasalRateStart(*datum.Last().Start+1))
	return datum
}

func CloneBasalRateStartArray(datumArray *pump.BasalRateStartArray) *pump.BasalRateStartArray {
	if datumArray == nil {
		return nil
	}
	clone := pump.NewBasalRateStartArray()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneBasalRateStart(datum))
	}
	return clone
}

func NewBasalRateStartArrayMap() *pump.BasalRateStartArrayMap {
	datum := pump.NewBasalRateStartArrayMap()
	datum.Set(dataTypesBasalTest.RandomScheduleName(), NewBasalRateStartArray())
	return datum
}

func CloneBasalRateStartArrayMap(datumArrayMap *pump.BasalRateStartArrayMap) *pump.BasalRateStartArrayMap {
	if datumArrayMap == nil {
		return nil
	}
	clone := pump.NewBasalRateStartArrayMap()
	for datumName, datumArray := range *datumArrayMap {
		clone.Set(datumName, CloneBasalRateStartArray(datumArray))
	}
	return clone
}
