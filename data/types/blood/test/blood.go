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
	datum.Base = *dataTypesTest.RandomBase()
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
	clone.Units = pointer.CloneString(datum.Units)
	clone.RawUnits = pointer.CloneString(datum.RawUnits)
	clone.Value = pointer.CloneFloat64(datum.Value)
	clone.RawValue = pointer.CloneFloat64(datum.RawValue)
	return clone
}
