package test

import (
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	testDataTypesBlood "github.com/tidepool-org/platform/data/types/blood/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewGlucose(units *string) *glucose.Glucose {
	datum := &glucose.Glucose{}
	datum.Blood = *testDataTypesBlood.NewBlood()
	datum.Units = units
	datum.Value = pointer.Float64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
	return datum
}

func CloneGlucose(datum *glucose.Glucose) *glucose.Glucose {
	if datum == nil {
		return nil
	}
	clone := &glucose.Glucose{}
	clone.Blood = *testDataTypesBlood.CloneBlood(&datum.Blood)
	return clone
}
