package test

import (
	"github.com/tidepool-org/platform/data/types/bolus/normal"
	dataTypesBolusTest "github.com/tidepool-org/platform/data/types/bolus/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewNormal() *normal.Normal {
	datum := normal.New()
	datum.Bolus = *dataTypesBolusTest.NewBolus()
	datum.SubType = "normal"
	datum.Normal = pointer.FromFloat64(test.RandomFloat64FromRange(normal.NormalMinimum, normal.NormalMaximum))
	datum.NormalExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Normal, normal.NormalMaximum))
	return datum
}

func CloneNormal(datum *normal.Normal) *normal.Normal {
	if datum == nil {
		return nil
	}
	clone := normal.New()
	clone.Bolus = *dataTypesBolusTest.CloneBolus(&datum.Bolus)
	clone.Normal = pointer.CloneFloat64(datum.Normal)
	clone.NormalExpected = pointer.CloneFloat64(datum.NormalExpected)
	return clone
}
