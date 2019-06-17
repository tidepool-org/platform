package test

import (
	"math"

	"github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewCompound(compoundArrayDepthLimit int) *insulin.Compound {
	datum := insulin.NewCompound()
	datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(0.0, math.MaxFloat64))
	datum.Formulation = NewFormulation(compoundArrayDepthLimit)
	return datum
}

func CloneCompound(datum *insulin.Compound) *insulin.Compound {
	if datum == nil {
		return nil
	}
	clone := insulin.NewCompound()
	clone.Amount = pointer.CloneFloat64(datum.Amount)
	clone.Formulation = CloneFormulation(datum.Formulation)
	return clone
}

func NewCompoundArray(compoundArrayDepthLimit int) *insulin.CompoundArray {
	if compoundArrayDepthLimit--; compoundArrayDepthLimit <= 0 {
		return nil
	}
	datum := insulin.NewCompoundArray()
	for count := 0; count < test.RandomIntFromRange(1, 3); count++ {
		*datum = append(*datum, NewCompound(compoundArrayDepthLimit))
	}
	return datum
}

func CloneCompoundArray(datumArray *insulin.CompoundArray) *insulin.CompoundArray {
	if datumArray == nil {
		return nil
	}
	clone := insulin.NewCompoundArray()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneCompound(datum))
	}
	return clone
}
