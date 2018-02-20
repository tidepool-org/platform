package test

import (
	"math/rand"

	"github.com/onsi/gomega"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewTarget(units *string) *dataBloodGlucose.Target {
	datum := dataBloodGlucose.NewTarget()
	switch rand.Intn(4) {
	case 0:
		datum.Target = pointer.Float64(test.RandomFloat64FromRange(dataBloodGlucose.TargetRangeForUnits(units)))
		datum.Range = pointer.Float64(test.RandomFloat64FromRange(dataBloodGlucose.RangeRangeForUnits(*datum.Target, units)))
	case 1:
		datum.Target = pointer.Float64(test.RandomFloat64FromRange(dataBloodGlucose.TargetRangeForUnits(units)))
		datum.High = pointer.Float64(test.RandomFloat64FromRange(dataBloodGlucose.HighRangeForUnits(*datum.Target, units)))
	case 2:
		datum.Target = pointer.Float64(test.RandomFloat64FromRange(dataBloodGlucose.TargetRangeForUnits(units)))
	case 3:
		datum.Low = pointer.Float64(test.RandomFloat64FromRange(dataBloodGlucose.LowRangeForUnits(units)))
		datum.High = pointer.Float64(test.RandomFloat64FromRange(dataBloodGlucose.HighRangeForUnits(*datum.Low, units)))
	}
	return datum
}

func CloneTarget(datum *dataBloodGlucose.Target) *dataBloodGlucose.Target {
	clone := dataBloodGlucose.NewTarget()
	clone.High = test.CloneFloat64(datum.High)
	clone.Low = test.CloneFloat64(datum.Low)
	clone.Range = test.CloneFloat64(datum.Range)
	clone.Target = test.CloneFloat64(datum.Target)
	return clone
}

func ExpectNormalizedUnits(value *string, expectedValue *string) {
	if expectedValue != nil {
		gomega.Expect(value).ToNot(gomega.BeNil())
		gomega.Expect(*value).To(gomega.Equal(*dataBloodGlucose.NormalizeUnits(expectedValue)))
		*expectedValue = *value
	} else {
		gomega.Expect(value).To(gomega.BeNil())
	}
}

func ExpectNormalizedValue(value *float64, expectedValue *float64, units *string) {
	if expectedValue != nil {
		gomega.Expect(value).ToNot(gomega.BeNil())
		gomega.Expect(*value).To(gomega.Equal(*dataBloodGlucose.NormalizeValueForUnits(expectedValue, units)))
		*expectedValue = *value
	} else {
		gomega.Expect(value).To(gomega.BeNil())
	}
}

func ExpectNormalizedTarget(datum *dataBloodGlucose.Target, expectedDatum *dataBloodGlucose.Target, units *string) {
	gomega.Expect(datum).ToNot(gomega.BeNil())
	gomega.Expect(expectedDatum).ToNot(gomega.BeNil())
	ExpectNormalizedValue(datum.High, expectedDatum.High, units)
	ExpectNormalizedValue(datum.Low, expectedDatum.Low, units)
	ExpectNormalizedValue(datum.Range, expectedDatum.Range, units)
	ExpectNormalizedValue(datum.Target, expectedDatum.Target, units)
}
