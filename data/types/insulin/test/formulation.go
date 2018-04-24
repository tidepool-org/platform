package test

import (
	"github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewFormulation(compoundArrayDepth int) *insulin.Formulation {
	simple := test.RandomBool()
	datum := insulin.NewFormulation()
	if !simple {
		datum.Compounds = NewCompoundArray(compoundArrayDepth)
	}
	datum.Name = pointer.String(test.NewText(1, 100))
	if simple {
		datum.Simple = NewSimple()
	}
	return datum
}

func CloneFormulation(datum *insulin.Formulation) *insulin.Formulation {
	if datum == nil {
		return nil
	}
	clone := insulin.NewFormulation()
	clone.Compounds = CloneCompoundArray(datum.Compounds)
	clone.Name = test.CloneString(datum.Name)
	clone.Simple = CloneSimple(datum.Simple)
	return clone
}
