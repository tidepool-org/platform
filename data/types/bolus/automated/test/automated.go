package test

import (
	dataTypesBolusAutomated "github.com/tidepool-org/platform/data/types/bolus/automated"
	dataTypesBolusTest "github.com/tidepool-org/platform/data/types/bolus/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomAutomated() *dataTypesBolusAutomated.Automated {
	datum := randomAutomated()
	datum.Bolus = *dataTypesBolusTest.RandomBolus()
	datum.SubType = "automated"
	return datum
}

func RandomAutomatedForParser() *dataTypesBolusAutomated.Automated {
	datum := randomAutomated()
	datum.Bolus = *dataTypesBolusTest.RandomBolusForParser()
	datum.SubType = "automated"
	return datum
}

func randomAutomated() *dataTypesBolusAutomated.Automated {
	datum := dataTypesBolusAutomated.New()
	datum.Normal = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesBolusAutomated.NormalMinimum, dataTypesBolusAutomated.NormalMaximum))
	datum.NormalExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Normal, dataTypesBolusAutomated.NormalMaximum))
	return datum
}

func CloneAutomated(datum *dataTypesBolusAutomated.Automated) *dataTypesBolusAutomated.Automated {
	if datum == nil {
		return nil
	}
	clone := dataTypesBolusAutomated.New()
	clone.Bolus = *dataTypesBolusTest.CloneBolus(&datum.Bolus)
	clone.Normal = pointer.CloneFloat64(datum.Normal)
	clone.NormalExpected = pointer.CloneFloat64(datum.NormalExpected)
	return clone
}

func NewObjectFromAutomated(datum *dataTypesBolusAutomated.Automated, objectFormat test.ObjectFormat) map[string]interface{} {
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
