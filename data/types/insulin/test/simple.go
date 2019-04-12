package test

import (
	"github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewSimple() *insulin.Simple {
	datum := insulin.NewSimple()
	datum.ActingType = pointer.FromString(test.RandomStringFromArray(insulin.SimpleActingTypes()))
	datum.Brand = pointer.FromString(test.RandomStringFromRange(1, 100))
	datum.Concentration = NewConcentration()
	return datum
}

func CloneSimple(datum *insulin.Simple) *insulin.Simple {
	if datum == nil {
		return nil
	}
	clone := insulin.NewSimple()
	clone.ActingType = pointer.CloneString(datum.ActingType)
	clone.Brand = pointer.CloneString(datum.Brand)
	clone.Concentration = CloneConcentration(datum.Concentration)
	return clone
}
