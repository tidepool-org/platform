package test

import (
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	dataTypesBloodTest "github.com/tidepool-org/platform/data/types/blood/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewGlucose(units *string) *glucose.Glucose {
	datum := &glucose.Glucose{}
	datum.Blood = *dataTypesBloodTest.NewBlood()
	datum.Units = units
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
	return datum
}

func CloneGlucose(datum *glucose.Glucose) *glucose.Glucose {
	if datum == nil {
		return nil
	}
	clone := &glucose.Glucose{}
	clone.Blood = *dataTypesBloodTest.CloneBlood(&datum.Blood)
	return clone
}
