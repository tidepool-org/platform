package test

import (
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataTypesBasalTest "github.com/tidepool-org/platform/data/types/basal/test"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewInsulinSensitivityStart(units *string, startMinimum int) *pump.InsulinSensitivityStart {
	datum := pump.NewInsulinSensitivityStart()
	datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
	datum.Start = pointer.FromInt(test.RandomIntFromRange(pump.InsulinSensitivityStartStartMinimum, pump.InsulinSensitivityStartStartMaximum))
	if startMinimum == pump.InsulinSensitivityStartStartMinimum {
		datum.Start = pointer.FromInt(pump.InsulinSensitivityStartStartMinimum)
	} else {
		datum.Start = pointer.FromInt(test.RandomIntFromRange(startMinimum, pump.InsulinSensitivityStartStartMaximum))
	}
	return datum
}

func CloneInsulinSensitivityStart(datum *pump.InsulinSensitivityStart) *pump.InsulinSensitivityStart {
	if datum == nil {
		return nil
	}
	clone := pump.NewInsulinSensitivityStart()
	clone.Amount = pointer.CloneFloat64(datum.Amount)
	clone.Start = pointer.CloneInt(datum.Start)
	return clone
}

func NewInsulinSensitivityStartArray(units *string) *pump.InsulinSensitivityStartArray {
	datum := pump.NewInsulinSensitivityStartArray()
	*datum = append(*datum, NewInsulinSensitivityStart(units, pump.InsulinSensitivityStartStartMinimum))
	*datum = append(*datum, NewInsulinSensitivityStart(units, *datum.Last().Start+1))
	*datum = append(*datum, NewInsulinSensitivityStart(units, *datum.Last().Start+1))
	return datum
}

func CloneInsulinSensitivityStartArray(datumArray *pump.InsulinSensitivityStartArray) *pump.InsulinSensitivityStartArray {
	if datumArray == nil {
		return nil
	}
	clone := pump.NewInsulinSensitivityStartArray()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneInsulinSensitivityStart(datum))
	}
	return clone
}

func NewInsulinSensitivityStartArrayMap(units *string) *pump.InsulinSensitivityStartArrayMap {
	datum := pump.NewInsulinSensitivityStartArrayMap()
	datum.Set(dataTypesBasalTest.RandomScheduleName(), NewInsulinSensitivityStartArray(units))
	return datum
}

func CloneInsulinSensitivityStartArrayMap(datumArrayMap *pump.InsulinSensitivityStartArrayMap) *pump.InsulinSensitivityStartArrayMap {
	if datumArrayMap == nil {
		return nil
	}
	clone := pump.NewInsulinSensitivityStartArrayMap()
	for datumName, datumArray := range *datumArrayMap {
		clone.Set(datumName, CloneInsulinSensitivityStartArray(datumArray))
	}
	return clone
}
