package test

import (
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/basal/scheduled"
	testDataTypesBasal "github.com/tidepool-org/platform/data/types/basal/test"
	testDataTypesInsulin "github.com/tidepool-org/platform/data/types/insulin/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewSuppressedScheduled() *scheduled.SuppressedScheduled {
	datum := scheduled.NewSuppressedScheduled()
	datum.Annotations = testData.NewBlobArray()
	datum.InsulinFormulation = testDataTypesInsulin.NewFormulation(3)
	datum.Rate = pointer.Float64(test.RandomFloat64FromRange(scheduled.RateMinimum, scheduled.RateMaximum))
	datum.ScheduleName = pointer.String(testDataTypesBasal.NewScheduleName())
	return datum
}

func CloneSuppressedScheduled(datum *scheduled.SuppressedScheduled) *scheduled.SuppressedScheduled {
	if datum == nil {
		return nil
	}
	clone := scheduled.NewSuppressedScheduled()
	clone.Type = datum.Type
	clone.DeliveryType = datum.DeliveryType
	clone.Annotations = testData.CloneBlobArray(datum.Annotations)
	clone.InsulinFormulation = testDataTypesInsulin.CloneFormulation(datum.InsulinFormulation)
	clone.Rate = test.CloneFloat64(datum.Rate)
	clone.ScheduleName = test.CloneString(datum.ScheduleName)
	return clone
}
