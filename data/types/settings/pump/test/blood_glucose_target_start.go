package test

import (
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomBloodGlucoseTargetStart(units *string, startMinimum int) *dataTypesSettingsPump.BloodGlucoseTargetStart {
	datum := dataTypesSettingsPump.NewBloodGlucoseTargetStart()
	datum.Target = *dataBloodGlucoseTest.NewTarget(units)
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
