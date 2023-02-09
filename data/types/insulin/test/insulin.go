package test

import (
	"github.com/tidepool-org/platform/data/types/insulin"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewInsulin() *insulin.Insulin {
	datum := insulin.New()
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "insulin"
	datum.Dose = NewDose()
	datum.Formulation = RandomFormulation(3)
	datum.Site = pointer.FromString(test.RandomStringFromRange(1, 100))
	return datum
}

func CloneInsulin(datum *insulin.Insulin) *insulin.Insulin {
	if datum == nil {
		return nil
	}
	clone := insulin.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.Dose = CloneDose(datum.Dose)
	clone.Formulation = CloneFormulation(datum.Formulation)
	clone.Site = pointer.CloneString(datum.Site)
	return clone
}
