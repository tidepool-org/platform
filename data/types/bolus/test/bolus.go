package test

import (
	dataTypesBolus "github.com/tidepool-org/platform/data/types/bolus"
	dataTypesInsulinTest "github.com/tidepool-org/platform/data/types/insulin/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomBolus() *dataTypesBolus.Bolus {
	datum := randomBolus()
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "bolus"
	return datum
}

func RandomBolusForParser() *dataTypesBolus.Bolus {
	datum := randomBolus()
	datum.Base = *dataTypesTest.RandomBaseForParser()
	datum.Type = "bolus"
	return datum
}

func randomBolus() *dataTypesBolus.Bolus {
	datum := &dataTypesBolus.Bolus{}
	datum.SubType = dataTypesTest.NewType()
	datum.InsulinFormulation = dataTypesInsulinTest.RandomFormulation(3)
	datum.DeliveryContext = pointer.FromString(test.RandomStringFromArray(dataTypesBolus.DeliveryContexts()))
	return datum
}

func CloneBolus(datum *dataTypesBolus.Bolus) *dataTypesBolus.Bolus {
	if datum == nil {
		return nil
	}
	clone := &dataTypesBolus.Bolus{}
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.SubType = datum.SubType
	clone.InsulinFormulation = dataTypesInsulinTest.CloneFormulation(datum.InsulinFormulation)
	clone.DeliveryContext = pointer.CloneString(datum.DeliveryContext)
	return clone
}

func NewObjectFromBolus(datum *dataTypesBolus.Bolus, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := dataTypesTest.NewObjectFromBase(&datum.Base, objectFormat)
	object["subType"] = test.NewObjectFromString(datum.SubType, objectFormat)
	if datum.InsulinFormulation != nil {
		object["insulinFormulation"] = dataTypesInsulinTest.NewObjectFromFormulation(datum.InsulinFormulation, objectFormat)
	}
	if datum.DeliveryContext != nil {
		object["deliveryContext"] = test.NewObjectFromString(*datum.DeliveryContext, objectFormat)
	}
	return object
}
