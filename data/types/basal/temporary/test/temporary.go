package test

import (
	dataTypesBasalAutomated "github.com/tidepool-org/platform/data/types/basal/automated"
	dataTypesBasalAutomatedTest "github.com/tidepool-org/platform/data/types/basal/automated/test"
	dataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled"
	dataTypesBasalScheduledTest "github.com/tidepool-org/platform/data/types/basal/scheduled/test"
	dataTypesBasalTemporary "github.com/tidepool-org/platform/data/types/basal/temporary"
	dataTypesInsulinTest "github.com/tidepool-org/platform/data/types/insulin/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomSuppressedTemporary(suppressed dataTypesBasalTemporary.Suppressed) *dataTypesBasalTemporary.SuppressedTemporary {
	datum := dataTypesBasalTemporary.NewSuppressedTemporary()
	datum.Annotations = metadataTest.RandomMetadataArray()
	datum.InsulinFormulation = dataTypesInsulinTest.RandomFormulation(3)
	datum.Percent = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesBasalTemporary.PercentMinimum, dataTypesBasalTemporary.PercentMaximum))
	datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesBasalTemporary.RateMinimum, dataTypesBasalTemporary.RateMaximum))
	datum.Suppressed = suppressed
	return datum
}

func CloneSuppressedTemporary(datum *dataTypesBasalTemporary.SuppressedTemporary) *dataTypesBasalTemporary.SuppressedTemporary {
	if datum == nil {
		return nil
	}
	clone := dataTypesBasalTemporary.NewSuppressedTemporary()
	clone.Type = datum.Type
	clone.DeliveryType = datum.DeliveryType
	clone.Annotations = metadataTest.CloneMetadataArray(datum.Annotations)
	clone.InsulinFormulation = dataTypesInsulinTest.CloneFormulation(datum.InsulinFormulation)
	clone.Percent = pointer.CloneFloat64(datum.Percent)
	clone.Rate = pointer.CloneFloat64(datum.Rate)
	if datum.Suppressed != nil {
		switch suppressed := datum.Suppressed.(type) {
		case *dataTypesBasalAutomated.SuppressedAutomated:
			clone.Suppressed = dataTypesBasalAutomatedTest.CloneSuppressedAutomated(suppressed)
		case *dataTypesBasalScheduled.SuppressedScheduled:
			clone.Suppressed = dataTypesBasalScheduledTest.CloneSuppressedScheduled(suppressed)
		}
	}
	return clone
}

func NewObjectFromSuppressedTemporary(datum *dataTypesBasalTemporary.SuppressedTemporary, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Type != nil {
		object["type"] = test.NewObjectFromString(*datum.Type, objectFormat)
	}
	if datum.DeliveryType != nil {
		object["deliveryType"] = test.NewObjectFromString(*datum.DeliveryType, objectFormat)
	}
	if datum.Annotations != nil {
		object["annotations"] = metadataTest.NewArrayFromMetadataArray(datum.Annotations, objectFormat)
	}
	if datum.InsulinFormulation != nil {
		object["insulinFormulation"] = dataTypesInsulinTest.NewObjectFromFormulation(datum.InsulinFormulation, objectFormat)
	}
	if datum.Percent != nil {
		object["percent"] = test.NewObjectFromFloat64(*datum.Percent, objectFormat)
	}
	if datum.Rate != nil {
		object["rate"] = test.NewObjectFromFloat64(*datum.Rate, objectFormat)
	}
	if datum.Suppressed != nil {
		switch suppressed := datum.Suppressed.(type) {
		case *dataTypesBasalAutomated.SuppressedAutomated:
			object["suppressed"] = dataTypesBasalAutomatedTest.NewObjectFromSuppressedAutomated(suppressed, objectFormat)
		case *dataTypesBasalScheduled.SuppressedScheduled:
			object["suppressed"] = dataTypesBasalScheduledTest.NewObjectFromSuppressedScheduled(suppressed, objectFormat)
		}
	}
	return object
}
