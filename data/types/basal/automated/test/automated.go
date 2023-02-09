package test

import (
	dataTypesBasalAutomated "github.com/tidepool-org/platform/data/types/basal/automated"
	dataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled"
	dataTypesBasalScheduledTest "github.com/tidepool-org/platform/data/types/basal/scheduled/test"
	dataTypesBasalTest "github.com/tidepool-org/platform/data/types/basal/test"
	dataTypesInsulinTest "github.com/tidepool-org/platform/data/types/insulin/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomAutomated() *dataTypesBasalAutomated.Automated {
	datum := randomAutomated()
	datum.Basal = *dataTypesBasalTest.RandomBasal()
	datum.DeliveryType = "automated"
	return datum
}

func RandomAutomatedForParser() *dataTypesBasalAutomated.Automated {
	datum := randomAutomated()
	datum.Basal = *dataTypesBasalTest.RandomBasalForParser()
	datum.DeliveryType = "automated"
	return datum
}

func randomAutomated() *dataTypesBasalAutomated.Automated {
	datum := dataTypesBasalAutomated.New()
	datum.Duration = pointer.FromInt(test.RandomIntFromRange(dataTypesBasalAutomated.DurationMinimum, dataTypesBasalAutomated.DurationMaximum))
	datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBasalAutomated.DurationMaximum))
	datum.InsulinFormulation = dataTypesInsulinTest.RandomFormulation(3)
	datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesBasalAutomated.RateMinimum, dataTypesBasalAutomated.RateMaximum))
	datum.ScheduleName = pointer.FromString(dataTypesBasalTest.RandomScheduleName())
	datum.Suppressed = dataTypesBasalScheduledTest.RandomSuppressedScheduled()
	return datum
}

func CloneAutomated(datum *dataTypesBasalAutomated.Automated) *dataTypesBasalAutomated.Automated {
	if datum == nil {
		return nil
	}
	clone := dataTypesBasalAutomated.New()
	clone.Basal = *dataTypesBasalTest.CloneBasal(&datum.Basal)
	clone.Duration = pointer.CloneInt(datum.Duration)
	clone.DurationExpected = pointer.CloneInt(datum.DurationExpected)
	clone.InsulinFormulation = dataTypesInsulinTest.CloneFormulation(datum.InsulinFormulation)
	clone.Rate = pointer.CloneFloat64(datum.Rate)
	clone.ScheduleName = pointer.CloneString(datum.ScheduleName)
	if datum.Suppressed != nil {
		switch suppressed := datum.Suppressed.(type) {
		case *dataTypesBasalScheduled.SuppressedScheduled:
			clone.Suppressed = dataTypesBasalScheduledTest.CloneSuppressedScheduled(suppressed)
		}
	}
	return clone
}

func NewObjectFromAutomated(datum *dataTypesBasalAutomated.Automated, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := dataTypesBasalTest.NewObjectFromBasal(&datum.Basal, objectFormat)
	if datum.Duration != nil {
		object["duration"] = test.NewObjectFromInt(*datum.Duration, objectFormat)
	}
	if datum.DurationExpected != nil {
		object["expectedDuration"] = test.NewObjectFromInt(*datum.DurationExpected, objectFormat)
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
	if datum.Suppressed != nil {
		switch suppressed := datum.Suppressed.(type) {
		case *dataTypesBasalScheduled.SuppressedScheduled:
			object["suppressed"] = dataTypesBasalScheduledTest.NewObjectFromSuppressedScheduled(suppressed, objectFormat)
		}
	}
	return object
}

func RandomSuppressedAutomated() *dataTypesBasalAutomated.SuppressedAutomated {
	datum := dataTypesBasalAutomated.NewSuppressedAutomated()
	datum.Annotations = metadataTest.RandomMetadataArray()
	datum.InsulinFormulation = dataTypesInsulinTest.RandomFormulation(3)
	datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesBasalAutomated.RateMinimum, dataTypesBasalAutomated.RateMaximum))
	datum.ScheduleName = pointer.FromString(dataTypesBasalTest.RandomScheduleName())
	datum.Suppressed = dataTypesBasalScheduledTest.RandomSuppressedScheduled()
	return datum
}

func CloneSuppressedAutomated(datum *dataTypesBasalAutomated.SuppressedAutomated) *dataTypesBasalAutomated.SuppressedAutomated {
	if datum == nil {
		return nil
	}
	clone := dataTypesBasalAutomated.NewSuppressedAutomated()
	clone.Type = datum.Type
	clone.DeliveryType = datum.DeliveryType
	clone.Annotations = metadataTest.CloneMetadataArray(datum.Annotations)
	clone.InsulinFormulation = dataTypesInsulinTest.CloneFormulation(datum.InsulinFormulation)
	clone.Rate = pointer.CloneFloat64(datum.Rate)
	clone.ScheduleName = pointer.CloneString(datum.ScheduleName)
	if datum.Suppressed != nil {
		switch suppressed := datum.Suppressed.(type) {
		case *dataTypesBasalScheduled.SuppressedScheduled:
			clone.Suppressed = dataTypesBasalScheduledTest.CloneSuppressedScheduled(suppressed)
		}
	}
	return clone
}

func NewObjectFromSuppressedAutomated(datum *dataTypesBasalAutomated.SuppressedAutomated, objectFormat test.ObjectFormat) map[string]interface{} {
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
	if datum.Suppressed != nil {
		switch suppressed := datum.Suppressed.(type) {
		case *dataTypesBasalScheduled.SuppressedScheduled:
			object["suppressed"] = dataTypesBasalScheduledTest.NewObjectFromSuppressedScheduled(suppressed, objectFormat)
		}
	}
	return object
}
