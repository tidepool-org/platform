package test

import (
	dataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled"
	dataTypesBasalTest "github.com/tidepool-org/platform/data/types/basal/test"
	dataTypesInsulinTest "github.com/tidepool-org/platform/data/types/insulin/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomSuppressedScheduled() *dataTypesBasalScheduled.SuppressedScheduled {
	datum := dataTypesBasalScheduled.NewSuppressedScheduled()
	datum.Annotations = metadataTest.RandomMetadataArray()
	datum.InsulinFormulation = dataTypesInsulinTest.RandomFormulation(3)
	datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesBasalScheduled.RateMinimum, dataTypesBasalScheduled.RateMaximum))
	datum.ScheduleName = pointer.FromString(dataTypesBasalTest.RandomScheduleName())
	return datum
}

func CloneSuppressedScheduled(datum *dataTypesBasalScheduled.SuppressedScheduled) *dataTypesBasalScheduled.SuppressedScheduled {
	if datum == nil {
		return nil
	}
	clone := dataTypesBasalScheduled.NewSuppressedScheduled()
	clone.Type = datum.Type
	clone.DeliveryType = datum.DeliveryType
	clone.Annotations = metadataTest.CloneMetadataArray(datum.Annotations)
	clone.InsulinFormulation = dataTypesInsulinTest.CloneFormulation(datum.InsulinFormulation)
	clone.Rate = pointer.CloneFloat64(datum.Rate)
	clone.ScheduleName = pointer.CloneString(datum.ScheduleName)
	return clone
}

func NewObjectFromSuppressedScheduled(datum *dataTypesBasalScheduled.SuppressedScheduled, objectFormat test.ObjectFormat) map[string]interface{} {
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
	if datum.Rate != nil {
		object["rate"] = test.NewObjectFromFloat64(*datum.Rate, objectFormat)
	}
	if datum.ScheduleName != nil {
		object["scheduleName"] = test.NewObjectFromString(*datum.ScheduleName, objectFormat)
	}
	return object
}
