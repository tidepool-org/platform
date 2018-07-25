package test

import (
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/basal/automated"
	testDataTypesBasal "github.com/tidepool-org/platform/data/types/basal/test"
	testDataTypesInsulin "github.com/tidepool-org/platform/data/types/insulin/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewSuppressedAutomated() *automated.SuppressedAutomated {
	datum := automated.NewSuppressedAutomated()
	datum.Annotations = testData.NewBlobArray()
	datum.InsulinFormulation = testDataTypesInsulin.NewFormulation(3)
	datum.Rate = pointer.Float64(test.RandomFloat64FromRange(automated.RateMinimum, automated.RateMaximum))
	datum.ScheduleName = pointer.String(testDataTypesBasal.NewScheduleName())
	return datum
}

func CloneSuppressedAutomated(datum *automated.SuppressedAutomated) *automated.SuppressedAutomated {
	if datum == nil {
		return nil
	}
	clone := automated.NewSuppressedAutomated()
	clone.Type = datum.Type
	clone.DeliveryType = datum.DeliveryType
	clone.Annotations = testData.CloneBlobArray(datum.Annotations)
	clone.InsulinFormulation = testDataTypesInsulin.CloneFormulation(datum.InsulinFormulation)
	clone.Rate = test.CloneFloat64(datum.Rate)
	clone.ScheduleName = test.CloneString(datum.ScheduleName)
	return clone
}
