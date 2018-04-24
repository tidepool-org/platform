package test

import (
	"github.com/tidepool-org/platform/data/types/bolus/normal"
	testDataTypesBolus "github.com/tidepool-org/platform/data/types/bolus/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewNormal() *normal.Normal {
	datum := normal.New()
	datum.Bolus = *testDataTypesBolus.NewBolus()
	datum.SubType = "normal"
	datum.Normal = pointer.Float64(test.RandomFloat64FromRange(normal.NormalMinimum, normal.NormalMaximum))
	datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, normal.NormalMaximum))
	return datum
}

func CloneNormal(datum *normal.Normal) *normal.Normal {
	if datum == nil {
		return nil
	}
	clone := normal.New()
	clone.Bolus = *testDataTypesBolus.CloneBolus(&datum.Bolus)
	clone.Normal = test.CloneFloat64(datum.Normal)
	clone.NormalExpected = test.CloneFloat64(datum.NormalExpected)
	return clone
}
