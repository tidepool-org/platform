package test

import (
	dataTypesBasalTest "github.com/tidepool-org/platform/data/types/basal/test"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewCarbohydrateRatioStart(startMinimum int) *pump.CarbohydrateRatioStart {
	datum := pump.NewCarbohydrateRatioStart()
	datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(pump.CarbohydrateRatioStartAmountMinimum, pump.CarbohydrateRatioStartAmountMaximum))
	if startMinimum == pump.CarbohydrateRatioStartStartMinimum {
		datum.Start = pointer.FromInt(pump.CarbohydrateRatioStartStartMinimum)
	} else {
		datum.Start = pointer.FromInt(test.RandomIntFromRange(startMinimum, pump.CarbohydrateRatioStartStartMaximum))
	}
	return datum
}

func CloneCarbohydrateRatioStart(datum *pump.CarbohydrateRatioStart) *pump.CarbohydrateRatioStart {
	if datum == nil {
		return nil
	}
	clone := pump.NewCarbohydrateRatioStart()
	clone.Amount = pointer.CloneFloat64(datum.Amount)
	clone.Start = pointer.CloneInt(datum.Start)
	return clone
}

func NewCarbohydrateRatioStartArray() *pump.CarbohydrateRatioStartArray {
	datum := pump.NewCarbohydrateRatioStartArray()
	*datum = append(*datum, NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum))
	*datum = append(*datum, NewCarbohydrateRatioStart(*datum.Last().Start+1))
	*datum = append(*datum, NewCarbohydrateRatioStart(*datum.Last().Start+1))
	return datum
}

func CloneCarbohydrateRatioStartArray(datumArray *pump.CarbohydrateRatioStartArray) *pump.CarbohydrateRatioStartArray {
	if datumArray == nil {
		return nil
	}
	clone := pump.NewCarbohydrateRatioStartArray()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneCarbohydrateRatioStart(datum))
	}
	return clone
}

func NewCarbohydrateRatioStartArrayMap() *pump.CarbohydrateRatioStartArrayMap {
	datum := pump.NewCarbohydrateRatioStartArrayMap()
	datum.Set(dataTypesBasalTest.RandomScheduleName(), NewCarbohydrateRatioStartArray())
	return datum
}

func CloneCarbohydrateRatioStartArrayMap(datumArrayMap *pump.CarbohydrateRatioStartArrayMap) *pump.CarbohydrateRatioStartArrayMap {
	if datumArrayMap == nil {
		return nil
	}
	clone := pump.NewCarbohydrateRatioStartArrayMap()
	for datumName, datumArray := range *datumArrayMap {
		clone.Set(datumName, CloneCarbohydrateRatioStartArray(datumArray))
	}
	return clone
}
