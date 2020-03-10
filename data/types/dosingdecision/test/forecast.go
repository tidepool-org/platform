package test

import (
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomForecast() *dataTypesDosingDecision.Forecast {
	datum := dataTypesDosingDecision.NewForecast()
	datum.Time = pointer.FromString(test.RandomTime().Format(dataTypesDosingDecision.TimeFormat))
	datum.Value = pointer.FromFloat64(test.RandomFloat64())
	return datum
}

func CloneForecast(datum *dataTypesDosingDecision.Forecast) *dataTypesDosingDecision.Forecast {
	if datum == nil {
		return nil
	}
	clone := dataTypesDosingDecision.NewForecast()
	clone.Time = pointer.CloneString(datum.Time)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}

func RandomForecastArray() *dataTypesDosingDecision.ForecastArray {
	datumArray := dataTypesDosingDecision.NewForecastArray()
	for count := test.RandomIntFromRange(1, 3); count > 0; count-- {
		*datumArray = append(*datumArray, RandomForecast())
	}
	return datumArray
}

func CloneForecastArray(datumArray *dataTypesDosingDecision.ForecastArray) *dataTypesDosingDecision.ForecastArray {
	if datumArray == nil {
		return nil
	}
	cloneArray := dataTypesDosingDecision.NewForecastArray()
	for _, datum := range *datumArray {
		*cloneArray = append(*cloneArray, CloneForecast(datum))
	}
	return cloneArray
}
