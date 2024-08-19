package test

import (
	"math"

	dataTypeInsulin "github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomCompound(compoundArrayDepthLimit int) *dataTypeInsulin.Compound {
	datum := dataTypeInsulin.NewCompound()
	datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(0.0, math.MaxFloat64))
	datum.Formulation = RandomFormulation(compoundArrayDepthLimit)
	return datum
}

func CloneCompound(datum *dataTypeInsulin.Compound) *dataTypeInsulin.Compound {
	if datum == nil {
		return nil
	}
	clone := dataTypeInsulin.NewCompound()
	clone.Amount = pointer.CloneFloat64(datum.Amount)
	clone.Formulation = CloneFormulation(datum.Formulation)
	return clone
}

func NewObjectFromCompound(datum *dataTypeInsulin.Compound, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Amount != nil {
		object["amount"] = test.NewObjectFromFloat64(*datum.Amount, objectFormat)
	}
	if datum.Formulation != nil {
		object["formulation"] = NewObjectFromFormulation(datum.Formulation, objectFormat)
	}
	return object
}

func RandomCompoundArray(compoundArrayDepthLimit int) *dataTypeInsulin.CompoundArray {
	if compoundArrayDepthLimit--; compoundArrayDepthLimit <= 0 {
		return nil
	}
	datum := dataTypeInsulin.NewCompoundArray()
	for count := 0; count < test.RandomIntFromRange(1, 3); count++ {
		*datum = append(*datum, RandomCompound(compoundArrayDepthLimit))
	}
	return datum
}

func CloneCompoundArray(datumArray *dataTypeInsulin.CompoundArray) *dataTypeInsulin.CompoundArray {
	if datumArray == nil {
		return nil
	}
	clone := dataTypeInsulin.NewCompoundArray()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneCompound(datum))
	}
	return clone
}

func NewArrayFromCompoundArray(datumArray *dataTypeInsulin.CompoundArray, objectFormat test.ObjectFormat) []interface{} {
	if datumArray == nil {
		return nil
	}
	array := []interface{}{}
	for _, datum := range *datumArray {
		array = append(array, NewObjectFromCompound(datum, objectFormat))
	}
	return array
}
