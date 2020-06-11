package test

import (
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataTypesBasalTest "github.com/tidepool-org/platform/data/types/basal/test"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"
)

func NewBloodGlucoseTargetStart(units *string, startMinimum int) *pump.BloodGlucoseTargetStart {
	datum := pump.NewBloodGlucoseTargetStart()
	datum.Target = *dataBloodGlucoseTest.NewTarget(units)
	if startMinimum == pump.BloodGlucoseTargetStartStartMinimum {
		datum.Start = pointer.FromInt(pump.BloodGlucoseTargetStartStartMinimum)
	} else {
		datum.Start = pointer.FromInt(test.RandomIntFromRange(startMinimum, pump.BloodGlucoseTargetStartStartMaximum))
	}
	return datum
}

func CloneBloodGlucoseTargetStart(datum *pump.BloodGlucoseTargetStart) *pump.BloodGlucoseTargetStart {
	if datum == nil {
		return nil
	}
	clone := pump.NewBloodGlucoseTargetStart()
	clone.Target = *dataBloodGlucoseTest.CloneTarget(&datum.Target)
	clone.Start = pointer.CloneInt(datum.Start)
	return clone
}

func NewBloodGlucoseTargetStartArray(units *string) *pump.BloodGlucoseTargetStartArray {
	datum := pump.NewBloodGlucoseTargetStartArray()
	*datum = append(*datum, NewBloodGlucoseTargetStart(units, pump.BloodGlucoseTargetStartStartMinimum))
	*datum = append(*datum, NewBloodGlucoseTargetStart(units, *datum.Last().Start+1))
	*datum = append(*datum, NewBloodGlucoseTargetStart(units, *datum.Last().Start+1))
	return datum
}

func CloneBloodGlucoseTargetStartArray(datumArray *pump.BloodGlucoseTargetStartArray) *pump.BloodGlucoseTargetStartArray {
	if datumArray == nil {
		return nil
	}
	clone := pump.NewBloodGlucoseTargetStartArray()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneBloodGlucoseTargetStart(datum))
	}
	return clone
}

func NewBloodGlucoseTargetStartArrayMap(units *string) *pump.BloodGlucoseTargetStartArrayMap {
	datum := pump.NewBloodGlucoseTargetStartArrayMap()
	datum.Set(dataTypesBasalTest.NewScheduleName(), NewBloodGlucoseTargetStartArray(units))
	return datum
}

func CloneBloodGlucoseTargetStartArrayMap(datumArrayMap *pump.BloodGlucoseTargetStartArrayMap) *pump.BloodGlucoseTargetStartArrayMap {
	if datumArrayMap == nil {
		return nil
	}
	clone := pump.NewBloodGlucoseTargetStartArrayMap()
	for datumName, datumArray := range *datumArrayMap {
		clone.Set(datumName, CloneBloodGlucoseTargetStartArray(datumArray))
	}
	return clone
}

type ValidatableWithUnitsAndStartMinimum interface {
	Validate(validator structure.Validator, units *string, startMinimum *int)
}

type ValidatableWithUnitsAndStartMinimumAdapter struct {
	validatableWithUnitsAndStartMinimum ValidatableWithUnitsAndStartMinimum
	units                               *string
	startMinimum                        *int
}

func NewValidatableWithUnitsAndStartMinimumAdapter(validatableWithUnitsAndStartMinimum ValidatableWithUnitsAndStartMinimum, units *string, startMinimum *int) *ValidatableWithUnitsAndStartMinimumAdapter {
	return &ValidatableWithUnitsAndStartMinimumAdapter{
		validatableWithUnitsAndStartMinimum: validatableWithUnitsAndStartMinimum,
		units:                               units,
		startMinimum:                        startMinimum,
	}
}

func (v *ValidatableWithUnitsAndStartMinimumAdapter) Validate(validator structure.Validator) {
	v.validatableWithUnitsAndStartMinimum.Validate(validator, v.units, v.startMinimum)
}
