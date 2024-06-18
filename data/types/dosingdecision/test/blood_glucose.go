package test

import (
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomBloodGlucose(units *string) *dataTypesDosingDecision.BloodGlucose {
	datum := dataTypesDosingDecision.NewBloodGlucose()
	datum.Time = pointer.FromTime(test.RandomTime())
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
	return datum
}

func CloneBloodGlucose(datum *dataTypesDosingDecision.BloodGlucose) *dataTypesDosingDecision.BloodGlucose {
	if datum == nil {
		return nil
	}
	clone := dataTypesDosingDecision.NewBloodGlucose()
	clone.Time = pointer.CloneTime(datum.Time)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}

func SetBloodGlucoseRaw(datum *dataTypesDosingDecision.BloodGlucose, normalized *dataTypesDosingDecision.BloodGlucose) {
	if datum != nil && normalized != nil {
		datum.RawValue = normalized.RawValue
	}

}

func NewObjectFromBloodGlucose(datum *dataTypesDosingDecision.BloodGlucose, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Time != nil {
		object["time"] = test.NewObjectFromTime(*datum.Time, objectFormat)
	}
	if datum.Value != nil {
		object["value"] = test.NewObjectFromFloat64(*datum.Value, objectFormat)
	}
	return object
}

func RandomBloodGlucoseArray(units *string) *dataTypesDosingDecision.BloodGlucoseArray {
	datum := dataTypesDosingDecision.NewBloodGlucoseArray()
	for count := 0; count < test.RandomIntFromRange(1, 3); count++ {
		*datum = append(*datum, RandomBloodGlucose(units))
	}
	return datum
}

func CloneBloodGlucoseArray(datumArray *dataTypesDosingDecision.BloodGlucoseArray) *dataTypesDosingDecision.BloodGlucoseArray {
	if datumArray == nil {
		return nil
	}
	clone := dataTypesDosingDecision.NewBloodGlucoseArray()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneBloodGlucose(datum))
	}
	return clone
}

func SetBloodGlucoseArrayRaw(datumArray *dataTypesDosingDecision.BloodGlucoseArray, normalizedArray *dataTypesDosingDecision.BloodGlucoseArray) {
	if datumArray != nil && normalizedArray != nil {
		for i, datum := range *datumArray {
			SetBloodGlucoseRaw(datum, (*normalizedArray)[i])
		}
	}
}

func NewArrayFromBloodGlucoseArray(datumArray *dataTypesDosingDecision.BloodGlucoseArray, objectFormat test.ObjectFormat) []interface{} {
	if datumArray == nil {
		return nil
	}
	array := []interface{}{}
	for _, datum := range *datumArray {
		array = append(array, NewObjectFromBloodGlucose(datum, objectFormat))
	}
	return array
}

func AnonymizeBloodGlucoseArray(datumArray *dataTypesDosingDecision.BloodGlucoseArray) []interface{} {
	array := []interface{}{}
	for _, datum := range *datumArray {
		array = append(array, datum)
	}
	return array
}
