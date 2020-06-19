package test

import (
	"github.com/tidepool-org/platform/data/types/bolus/pen"
	dataTypesBolusTest "github.com/tidepool-org/platform/data/types/bolus/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewPen() *pen.Pen {
	datum := pen.New()
	datum.Bolus = *dataTypesBolusTest.NewBolus()
	datum.SubType = "pen"
	datum.Normal = pointer.FromFloat64(test.RandomFloat64FromRange(pen.NormalMinimum, pen.NormalMaximum))
	return datum
}

func ClonePen(datum *pen.Pen) *pen.Pen {
	if datum == nil {
		return nil
	}
	clone := pen.New()
	clone.Bolus = *dataTypesBolusTest.CloneBolus(&datum.Bolus)
	clone.Normal = pointer.CloneFloat64(datum.Normal)
	return clone
}
