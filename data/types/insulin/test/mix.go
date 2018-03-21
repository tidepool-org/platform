package test

import (
	"math"

	"github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewMixElement() *insulin.MixElement {
	datum := insulin.NewMixElement()
	datum.Amount = pointer.Float64(test.RandomFloat64FromRange(0.0, math.MaxFloat64))
	datum.Formulation = NewFormulation()
	return datum
}

func CloneMixElement(datum *insulin.MixElement) *insulin.MixElement {
	if datum == nil {
		return nil
	}
	clone := insulin.NewMixElement()
	clone.Amount = test.CloneFloat64(datum.Amount)
	clone.Formulation = CloneFormulation(datum.Formulation)
	return clone
}

func NewMix() *insulin.Mix {
	datum := insulin.NewMix()
	*datum = append(*datum, NewMixElement())
	return datum
}

func CloneMix(datumArray *insulin.Mix) *insulin.Mix {
	if datumArray == nil {
		return nil
	}
	clone := insulin.NewMix()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneMixElement(datum))
	}
	return clone
}
