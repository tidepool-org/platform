package test

import (
	dataTypesBolusNormal "github.com/tidepool-org/platform/data/types/bolus/normal"
	dataTypesBolusTest "github.com/tidepool-org/platform/data/types/bolus/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomNormalFields() dataTypesBolusNormal.NormalFields {
	datum := dataTypesBolusNormal.NormalFields{}
	datum.Normal = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesBolusNormal.NormalMinimum, dataTypesBolusNormal.NormalMaximum))
	if test.RandomBool() {
		datum.NormalExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Normal, dataTypesBolusNormal.NormalMaximum))
	}
	return datum
}

func RandomNormal() *dataTypesBolusNormal.Normal {
	datum := dataTypesBolusNormal.New()
	datum.Bolus = *dataTypesBolusTest.RandomBolus()
	datum.SubType = dataTypesBolusNormal.SubType
	datum.NormalFields = RandomNormalFields()
	return datum
}

func RandomNormalForParser() *dataTypesBolusNormal.Normal {
	datum := dataTypesBolusNormal.New()
	datum.Bolus = *dataTypesBolusTest.RandomBolusForParser()
	datum.SubType = dataTypesBolusNormal.SubType
	datum.NormalFields = RandomNormalFields()
	return datum
}

func CloneNormal(datum *dataTypesBolusNormal.Normal) *dataTypesBolusNormal.Normal {
	if datum == nil {
		return nil
	}
	clone := dataTypesBolusNormal.New()
	clone.Bolus = *dataTypesBolusTest.CloneBolus(&datum.Bolus)
	clone.NormalFields = datum.NormalFields
	return clone
}

func NewObjectFromNormal(datum *dataTypesBolusNormal.Normal, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := dataTypesBolusTest.NewObjectFromBolus(&datum.Bolus, objectFormat)
	if datum.Normal != nil {
		object["normal"] = test.NewObjectFromFloat64(*datum.Normal, objectFormat)
	}
	if datum.NormalExpected != nil {
		object["expectedNormal"] = test.NewObjectFromFloat64(*datum.NormalExpected, objectFormat)
	}
	return object
}
