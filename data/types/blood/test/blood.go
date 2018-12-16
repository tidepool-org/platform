package test

import (
	"math"

	"github.com/tidepool-org/platform/data/types/blood"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewBlood() *blood.Blood {
	datum := &blood.Blood{}
	datum.Base = *dataTypesTest.NewBase()
	datum.Units = pointer.FromString(dataTypesTest.NewType())
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
	return datum
}

func CloneBlood(datum *blood.Blood) *blood.Blood {
	if datum == nil {
		return nil
	}
	clone := &blood.Blood{}
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.Units = test.CloneString(datum.Units)
	clone.Value = test.CloneFloat64(datum.Value)
	return clone
}
