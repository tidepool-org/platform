package test

import (
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/basal/automated"
	dataTypesBasalTest "github.com/tidepool-org/platform/data/types/basal/test"
	dataTypesInsulinTest "github.com/tidepool-org/platform/data/types/insulin/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewSuppressedAutomated() *automated.SuppressedAutomated {
	datum := automated.NewSuppressedAutomated()
	datum.Annotations = dataTest.NewBlobArray()
	datum.InsulinFormulation = dataTypesInsulinTest.NewFormulation(3)
	datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(automated.RateMinimum, automated.RateMaximum))
	datum.ScheduleName = pointer.FromString(dataTypesBasalTest.NewScheduleName())
	return datum
}

func CloneSuppressedAutomated(datum *automated.SuppressedAutomated) *automated.SuppressedAutomated {
	if datum == nil {
		return nil
	}
	clone := automated.NewSuppressedAutomated()
	clone.Type = datum.Type
	clone.DeliveryType = datum.DeliveryType
	clone.Annotations = dataTest.CloneBlobArray(datum.Annotations)
	clone.InsulinFormulation = dataTypesInsulinTest.CloneFormulation(datum.InsulinFormulation)
	clone.Rate = test.CloneFloat64(datum.Rate)
	clone.ScheduleName = test.CloneString(datum.ScheduleName)
	return clone
}
