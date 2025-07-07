package test

import (
	dataTypesBolusExtended "github.com/tidepool-org/platform/data/types/bolus/extended"
	dataTypesBolusTest "github.com/tidepool-org/platform/data/types/bolus/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomExtendedFields() dataTypesBolusExtended.ExtendedFields {
	datum := dataTypesBolusExtended.ExtendedFields{}
	datum.Duration = pointer.FromInt(test.RandomIntFromRange(dataTypesBolusExtended.DurationMinimum, dataTypesBolusExtended.DurationMaximum))
	datum.Extended = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesBolusExtended.ExtendedMinimum, dataTypesBolusExtended.ExtendedMaximum))
	if test.RandomBool() {
		datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
		datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
	}
	return datum
}

func RandomExtended() *dataTypesBolusExtended.Extended {
	datum := dataTypesBolusExtended.New()
	datum.Bolus = *dataTypesBolusTest.RandomBolus()
	datum.SubType = dataTypesBolusExtended.SubType
	datum.ExtendedFields = RandomExtendedFields()
	return datum
}

func RandomExtendedForParser() *dataTypesBolusExtended.Extended {
	datum := dataTypesBolusExtended.New()
	datum.Bolus = *dataTypesBolusTest.RandomBolusForParser()
	datum.SubType = dataTypesBolusExtended.SubType
	datum.ExtendedFields = RandomExtendedFields()
	return datum
}

func CloneExtended(datum *dataTypesBolusExtended.Extended) *dataTypesBolusExtended.Extended {
	if datum == nil {
		return nil
	}
	clone := dataTypesBolusExtended.New()
	clone.Bolus = *dataTypesBolusTest.CloneBolus(&datum.Bolus)
	clone.ExtendedFields = datum.ExtendedFields
	return clone
}

func NewObjectFromExtended(datum *dataTypesBolusExtended.Extended, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := dataTypesBolusTest.NewObjectFromBolus(&datum.Bolus, objectFormat)
	if datum.Duration != nil {
		object["duration"] = test.NewObjectFromInt(*datum.Duration, objectFormat)
	}
	if datum.DurationExpected != nil {
		object["expectedDuration"] = test.NewObjectFromInt(*datum.DurationExpected, objectFormat)
	}
	if datum.Extended != nil {
		object["extended"] = test.NewObjectFromFloat64(*datum.Extended, objectFormat)
	}
	if datum.ExtendedExpected != nil {
		object["expectedExtended"] = test.NewObjectFromFloat64(*datum.ExtendedExpected, objectFormat)
	}
	return object
}
