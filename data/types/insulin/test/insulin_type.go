package test

import (
	"github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/test"
)

func NewInsulinType() *insulin.InsulinType {
	datum := insulin.NewInsulinType()
	if test.RandomBool() {
		datum.Formulation = NewFormulation()
	} else {
		datum.Mix = NewMix()
	}
	return datum
}

func CloneInsulinType(datum *insulin.InsulinType) *insulin.InsulinType {
	if datum == nil {
		return nil
	}
	clone := insulin.NewInsulinType()
	clone.Formulation = CloneFormulation(datum.Formulation)
	clone.Mix = CloneMix(datum.Mix)
	return clone
}
