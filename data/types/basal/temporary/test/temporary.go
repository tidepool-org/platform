package test

import (
	testData "github.com/tidepool-org/platform/data/test"
	dataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled"
	testDataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled/test"
	"github.com/tidepool-org/platform/data/types/basal/temporary"
	testDataTypesInsulin "github.com/tidepool-org/platform/data/types/insulin/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewSuppressedTemporary(suppressed temporary.Suppressed) *temporary.SuppressedTemporary {
	datum := temporary.NewSuppressedTemporary()
	datum.Annotations = testData.NewBlobArray()
	datum.InsulinFormulation = testDataTypesInsulin.NewFormulation(3)
	datum.Percent = pointer.Float64(test.RandomFloat64FromRange(temporary.PercentMinimum, temporary.PercentMaximum))
	datum.Rate = pointer.Float64(test.RandomFloat64FromRange(temporary.RateMinimum, temporary.RateMaximum))
	datum.Suppressed = suppressed
	return datum
}

func CloneSuppressedTemporary(datum *temporary.SuppressedTemporary) *temporary.SuppressedTemporary {
	if datum == nil {
		return nil
	}
	clone := temporary.NewSuppressedTemporary()
	clone.Type = datum.Type
	clone.DeliveryType = datum.DeliveryType
	clone.Annotations = testData.CloneBlobArray(datum.Annotations)
	clone.InsulinFormulation = testDataTypesInsulin.CloneFormulation(datum.InsulinFormulation)
	clone.Percent = test.CloneFloat64(datum.Percent)
	clone.Rate = test.CloneFloat64(datum.Rate)
	if datum.Suppressed != nil {
		switch suppressed := datum.Suppressed.(type) {
		case *dataTypesBasalScheduled.SuppressedScheduled:
			clone.Suppressed = testDataTypesBasalScheduled.CloneSuppressedScheduled(suppressed)
		}
	}
	return clone
}
