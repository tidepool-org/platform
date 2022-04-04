package test

import (
	"github.com/tidepool-org/platform/data/types/bolus"
	dataTypesInsulinTest "github.com/tidepool-org/platform/data/types/insulin/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
)

func NewBolus() *bolus.Bolus {
	datum := &bolus.Bolus{}
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "bolus"
	datum.SubType = dataTypesTest.NewType()
	datum.InsulinFormulation = dataTypesInsulinTest.NewFormulation(3)
	return datum
}

func CloneBolus(datum *bolus.Bolus) *bolus.Bolus {
	if datum == nil {
		return nil
	}
	clone := &bolus.Bolus{}
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.SubType = datum.SubType
	clone.InsulinFormulation = dataTypesInsulinTest.CloneFormulation(datum.InsulinFormulation)
	return clone
}
