package test

import (
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/basal/scheduled"
	dataTypesBasalTest "github.com/tidepool-org/platform/data/types/basal/test"
	dataTypesInsulinTest "github.com/tidepool-org/platform/data/types/insulin/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewSuppressedScheduled() *scheduled.SuppressedScheduled {
	datum := scheduled.NewSuppressedScheduled()
	datum.Annotations = dataTest.NewBlobArray()
	datum.InsulinFormulation = dataTypesInsulinTest.NewFormulation(3)
	datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(scheduled.RateMinimum, scheduled.RateMaximum))
	datum.ScheduleName = pointer.FromString(dataTypesBasalTest.NewScheduleName())
	return datum
}

func CloneSuppressedScheduled(datum *scheduled.SuppressedScheduled) *scheduled.SuppressedScheduled {
	if datum == nil {
		return nil
	}
	clone := scheduled.NewSuppressedScheduled()
	clone.Type = datum.Type
	clone.DeliveryType = datum.DeliveryType
	clone.Annotations = dataTest.CloneBlobArray(datum.Annotations)
	clone.InsulinFormulation = dataTypesInsulinTest.CloneFormulation(datum.InsulinFormulation)
	clone.Rate = test.CloneFloat64(datum.Rate)
	clone.ScheduleName = test.CloneString(datum.ScheduleName)
	return clone
}
