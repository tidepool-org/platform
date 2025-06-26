package test

import (
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomForecastBloodGlucose(units *string) *dataTypesDosingDecision.ForecastBloodGlucose {
	datum := dataTypesDosingDecision.NewForecastBloodGlucose()
	datum.Time = pointer.FromTime(test.RandomTime())
	datum.Value = pointer.FromFloat64(test.RandomFloat64())
	return datum
}

func CloneForecastBloodGlucose(datum *dataTypesDosingDecision.ForecastBloodGlucose) *dataTypesDosingDecision.ForecastBloodGlucose {
	if datum == nil {
		return nil
	}
	clone := dataTypesDosingDecision.NewForecastBloodGlucose()
	clone.Time = pointer.CloneTime(datum.Time)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}

func NewObjectFromForecastBloodGlucose(datum *dataTypesDosingDecision.ForecastBloodGlucose, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.Time != nil {
		object["time"] = test.NewObjectFromTime(*datum.Time, objectFormat)
	}
	if datum.Value != nil {
		object["value"] = test.NewObjectFromFloat64(*datum.Value, objectFormat)
	}
	return object
}

func RandomForecastBloodGlucoseArray(units *string) *dataTypesDosingDecision.ForecastBloodGlucoseArray {
	datum := dataTypesDosingDecision.NewForecastBloodGlucoseArray()
	for count := 0; count < test.RandomIntFromRange(1, 3); count++ {
		*datum = append(*datum, RandomForecastBloodGlucose(units))
	}
	return datum
}

func CloneForecastBloodGlucoseArray(datumArray *dataTypesDosingDecision.ForecastBloodGlucoseArray) *dataTypesDosingDecision.ForecastBloodGlucoseArray {
	if datumArray == nil {
		return nil
	}
	clone := dataTypesDosingDecision.NewForecastBloodGlucoseArray()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneForecastBloodGlucose(datum))
	}
	return clone
}

func NewArrayFromForecastBloodGlucoseArray(datumArray *dataTypesDosingDecision.ForecastBloodGlucoseArray, objectFormat test.ObjectFormat) []any {
	if datumArray == nil {
		return nil
	}
	array := []any{}
	for _, datum := range *datumArray {
		array = append(array, NewObjectFromForecastBloodGlucose(datum, objectFormat))
	}
	return array
}

func AnonymizeForecastBloodGlucoseArray(datumArray *dataTypesDosingDecision.ForecastBloodGlucoseArray) []any {
	array := []any{}
	for _, datum := range *datumArray {
		array = append(array, datum)
	}
	return array
}
