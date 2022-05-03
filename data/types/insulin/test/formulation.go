package test

import (
	dataTypeInsulin "github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomFormulation(compoundArrayDepthLimit int) *dataTypeInsulin.Formulation {
	simple := test.RandomBool()
	datum := dataTypeInsulin.NewFormulation()
	if !simple {
		datum.Compounds = RandomCompoundArray(compoundArrayDepthLimit)
	}
	datum.Name = pointer.FromString(test.RandomStringFromRange(1, 100))
	if simple {
		datum.Simple = RandomSimple()
	}
	return datum
}

func CloneFormulation(datum *dataTypeInsulin.Formulation) *dataTypeInsulin.Formulation {
	if datum == nil {
		return nil
	}
	clone := dataTypeInsulin.NewFormulation()
	clone.Compounds = CloneCompoundArray(datum.Compounds)
	clone.Name = pointer.CloneString(datum.Name)
	clone.Simple = CloneSimple(datum.Simple)
	return clone
}

func NewObjectFromFormulation(datum *dataTypeInsulin.Formulation, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Compounds != nil {
		object["compounds"] = NewArrayFromCompoundArray(datum.Compounds, objectFormat)
	}
	if datum.Name != nil {
		object["name"] = test.NewObjectFromString(*datum.Name, objectFormat)
	}
	if datum.Simple != nil {
		object["simple"] = NewObjectFromSimple(datum.Simple, objectFormat)
	}
	return object
}
