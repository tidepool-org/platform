package test

import (
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataTypesBasalTest "github.com/tidepool-org/platform/data/types/basal/test"
	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"
)

func RandomBloodGlucoseTargetStart(units *string, startMinimum int) *dataTypesSettingsPump.BloodGlucoseTargetStart {
	datum := dataTypesSettingsPump.NewBloodGlucoseTargetStart()
	datum.Target = *dataBloodGlucoseTest.RandomTarget(units)
	if startMinimum == dataTypesSettingsPump.BloodGlucoseTargetStartStartMinimum {
		datum.Start = pointer.FromInt(dataTypesSettingsPump.BloodGlucoseTargetStartStartMinimum)
	} else {
		datum.Start = pointer.FromInt(test.RandomIntFromRange(startMinimum, dataTypesSettingsPump.BloodGlucoseTargetStartStartMaximum))
	}
	return datum
}

func CloneBloodGlucoseTargetStart(datum *dataTypesSettingsPump.BloodGlucoseTargetStart) *dataTypesSettingsPump.BloodGlucoseTargetStart {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsPump.NewBloodGlucoseTargetStart()
	clone.Target = *dataBloodGlucoseTest.CloneTarget(&datum.Target)
	clone.Start = pointer.CloneInt(datum.Start)
	return clone
}

func NewObjectFromBloodGlucoseTargetStart(datum *dataTypesSettingsPump.BloodGlucoseTargetStart, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := dataBloodGlucoseTest.NewObjectFromTarget(&datum.Target, objectFormat)
	if datum.Start != nil {
		object["start"] = test.NewObjectFromInt(*datum.Start, objectFormat)
	}
	return object
}

func RandomBloodGlucoseTargetStartArray(units *string) *dataTypesSettingsPump.BloodGlucoseTargetStartArray {
	startMinimum := dataTypesSettingsPump.BloodGlucoseTargetStartStartMinimum
	datumArray := dataTypesSettingsPump.NewBloodGlucoseTargetStartArray()
	for count := test.RandomIntFromRange(1, 3); count > 0; count-- {
		datum := RandomBloodGlucoseTargetStart(units, startMinimum)
		*datumArray = append(*datumArray, datum)
		startMinimum = *datum.Start + 1
	}
	return datumArray
}

func CloneBloodGlucoseTargetStartArray(datumArray *dataTypesSettingsPump.BloodGlucoseTargetStartArray) *dataTypesSettingsPump.BloodGlucoseTargetStartArray {
	if datumArray == nil {
		return nil
	}
	cloneArray := dataTypesSettingsPump.NewBloodGlucoseTargetStartArray()
	for _, datum := range *datumArray {
		*cloneArray = append(*cloneArray, CloneBloodGlucoseTargetStart(datum))
	}
	return cloneArray
}

func NewArrayFromBloodGlucoseTargetStartArray(datumArray *dataTypesSettingsPump.BloodGlucoseTargetStartArray, objectFormat test.ObjectFormat) []interface{} {
	if datumArray == nil {
		return nil
	}
	array := []interface{}{}
	for _, datum := range *datumArray {
		array = append(array, NewObjectFromBloodGlucoseTargetStart(datum, objectFormat))
	}
	return array
}

func AnonymizeBloodGlucoseTargetStartArray(datumArray *dataTypesSettingsPump.BloodGlucoseTargetStartArray) []interface{} {
	array := []interface{}{}
	for _, datum := range *datumArray {
		array = append(array, datum)
	}
	return array
}

func CloneBloodGlucoseTargetStartArrayMap(datumArrayMap *dataTypesSettingsPump.BloodGlucoseTargetStartArrayMap) *dataTypesSettingsPump.BloodGlucoseTargetStartArrayMap {
	if datumArrayMap == nil {
		return nil
	}
	clone := dataTypesSettingsPump.NewBloodGlucoseTargetStartArrayMap()
	for datumName, datumArray := range *datumArrayMap {
		clone.Set(datumName, CloneBloodGlucoseTargetStartArray(datumArray))
	}
	return clone
}

func NewObjectFromBloodGlucoseTargetStartArrayMap(datumArrayMap *dataTypesSettingsPump.BloodGlucoseTargetStartArrayMap, objectFormat test.ObjectFormat) map[string]interface{} {
	if datumArrayMap == nil {
		return nil
	}
	object := map[string]interface{}{}
	for datumName, datumArray := range *datumArrayMap {
		object[datumName] = NewArrayFromBloodGlucoseTargetStartArray(datumArray, objectFormat)
	}
	return object
}

func NewBloodGlucoseTargetStartArrayMap(units *string) *dataTypesSettingsPump.BloodGlucoseTargetStartArrayMap {
	datum := dataTypesSettingsPump.NewBloodGlucoseTargetStartArrayMap()
	datum.Set(dataTypesBasalTest.RandomScheduleName(), RandomBloodGlucoseTargetStartArray(units))
	return datum
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
