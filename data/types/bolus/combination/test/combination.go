package test

import (
	dataTypesBolusCombination "github.com/tidepool-org/platform/data/types/bolus/combination"
	dataTypesBolusExtendedTest "github.com/tidepool-org/platform/data/types/bolus/extended/test"
	dataTypesBolusNormalTest "github.com/tidepool-org/platform/data/types/bolus/normal/test"
	dataTypesBolusTest "github.com/tidepool-org/platform/data/types/bolus/test"
	"github.com/tidepool-org/platform/test"
)

func RandomCombinationFields() dataTypesBolusCombination.CombinationFields {
	datum := dataTypesBolusCombination.CombinationFields{}
	datum.ExtendedFields = dataTypesBolusExtendedTest.RandomExtendedFields()
	datum.NormalFields = dataTypesBolusNormalTest.RandomNormalFields()
	return datum
}

func RandomCombination() *dataTypesBolusCombination.Combination {
	datum := dataTypesBolusCombination.New()
	datum.Bolus = *dataTypesBolusTest.RandomBolus()
	datum.SubType = dataTypesBolusCombination.SubType
	datum.CombinationFields = RandomCombinationFields()
	return datum
}

func RandomCombinationForParser() *dataTypesBolusCombination.Combination {
	datum := dataTypesBolusCombination.New()
	datum.Bolus = *dataTypesBolusTest.RandomBolusForParser()
	datum.SubType = dataTypesBolusCombination.SubType
	datum.CombinationFields = RandomCombinationFields()
	return datum
}

func CloneCombination(datum *dataTypesBolusCombination.Combination) *dataTypesBolusCombination.Combination {
	if datum == nil {
		return nil
	}
	clone := dataTypesBolusCombination.New()
	clone.Bolus = *dataTypesBolusTest.CloneBolus(&datum.Bolus)
	clone.CombinationFields = datum.CombinationFields
	return clone
}

func NewObjectFromCombination(datum *dataTypesBolusCombination.Combination, objectFormat test.ObjectFormat) map[string]any {
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
	if datum.Normal != nil {
		object["normal"] = test.NewObjectFromFloat64(*datum.Normal, objectFormat)
	}
	if datum.NormalExpected != nil {
		object["expectedNormal"] = test.NewObjectFromFloat64(*datum.NormalExpected, objectFormat)
	}
	return object
}
