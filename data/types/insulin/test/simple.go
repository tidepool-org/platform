package test

import (
	"github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewSimple() *insulin.Simple {
	datum := insulin.NewSimple()
	datum.ActingType = pointer.String(test.RandomStringFromArray(insulin.SimpleActingTypes()))
	datum.Brand = pointer.String(test.NewText(1, 100))
	datum.Concentration = NewConcentration()
	return datum
}

func CloneSimple(datum *insulin.Simple) *insulin.Simple {
	if datum == nil {
		return nil
	}
	clone := insulin.NewSimple()
	clone.ActingType = test.CloneString(datum.ActingType)
	clone.Brand = test.CloneString(datum.Brand)
	clone.Concentration = CloneConcentration(datum.Concentration)
	return clone
}
