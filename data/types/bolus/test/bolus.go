package test

import (
	"github.com/tidepool-org/platform/data/types/bolus"
	testDataTypesInsulin "github.com/tidepool-org/platform/data/types/insulin/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
)

func NewBolus() *bolus.Bolus {
	datum := &bolus.Bolus{}
	datum.Base = *testDataTypes.NewBase()
	datum.Type = "bolus"
	datum.SubType = testDataTypes.NewType()
	datum.InsulinType = testDataTypesInsulin.NewInsulinType()
	return datum
}

func CloneBolus(datum *bolus.Bolus) *bolus.Bolus {
	if datum == nil {
		return nil
	}
	clone := &bolus.Bolus{}
	clone.Base = *testDataTypes.CloneBase(&datum.Base)
	clone.SubType = datum.SubType
	clone.InsulinType = testDataTypesInsulin.CloneInsulinType(datum.InsulinType)
	return clone
}
