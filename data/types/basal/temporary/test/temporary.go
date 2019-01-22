package test

import (
	dataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled"
	dataTypesBasalScheduledTest "github.com/tidepool-org/platform/data/types/basal/scheduled/test"
	"github.com/tidepool-org/platform/data/types/basal/temporary"
	dataTypesInsulinTest "github.com/tidepool-org/platform/data/types/insulin/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewSuppressedTemporary(suppressed temporary.Suppressed) *temporary.SuppressedTemporary {
	datum := temporary.NewSuppressedTemporary()
	datum.Annotations = metadataTest.RandomMetadataArray()
	datum.InsulinFormulation = dataTypesInsulinTest.NewFormulation(3)
	datum.Percent = pointer.FromFloat64(test.RandomFloat64FromRange(temporary.PercentMinimum, temporary.PercentMaximum))
	datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(temporary.RateMinimum, temporary.RateMaximum))
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
	clone.Annotations = metadataTest.CloneMetadataArray(datum.Annotations)
	clone.InsulinFormulation = dataTypesInsulinTest.CloneFormulation(datum.InsulinFormulation)
	clone.Percent = pointer.CloneFloat64(datum.Percent)
	clone.Rate = pointer.CloneFloat64(datum.Rate)
	if datum.Suppressed != nil {
		switch suppressed := datum.Suppressed.(type) {
		case *dataTypesBasalScheduled.SuppressedScheduled:
			clone.Suppressed = dataTypesBasalScheduledTest.CloneSuppressedScheduled(suppressed)
		}
	}
	return clone
}
