package test

import (
	"github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewFormulation() *insulin.Formulation {
	datum := insulin.NewFormulation()
	datum.ActingType = pointer.String(test.RandomStringFromArray(insulin.FormulationActingTypes()))
	datum.Brand = pointer.String(test.NewText(1, 100))
	datum.Concentration = NewConcentration()
	datum.Name = pointer.String(test.NewText(1, 100))
	return datum
}

func CloneFormulation(datum *insulin.Formulation) *insulin.Formulation {
	if datum == nil {
		return nil
	}
	clone := insulin.NewFormulation()
	clone.ActingType = test.CloneString(datum.ActingType)
	clone.Brand = test.CloneString(datum.Brand)
	clone.Concentration = CloneConcentration(datum.Concentration)
	clone.Name = test.CloneString(datum.Name)
	return clone
}
