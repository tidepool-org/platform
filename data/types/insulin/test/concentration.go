package test

import (
	"github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewConcentration() *insulin.Concentration {
	datum := insulin.NewConcentration()
	datum.Units = pointer.FromString(test.RandomStringFromArray(insulin.ConcentrationUnits()))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(insulin.ConcentrationValueRangeForUnits(datum.Units)))
	return datum
}

func CloneConcentration(datum *insulin.Concentration) *insulin.Concentration {
	if datum == nil {
		return nil
	}
	clone := insulin.NewConcentration()
	clone.Units = test.CloneString(datum.Units)
	clone.Value = test.CloneFloat64(datum.Value)
	return clone
}
