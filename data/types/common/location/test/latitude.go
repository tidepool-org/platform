package test

import (
	"github.com/tidepool-org/platform/data/types/common/location"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewLatitude() *location.Latitude {
	datum := location.NewLatitude()
	datum.Units = pointer.FromString("degrees")
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(-90.0, 90.0))
	return datum
}

func CloneLatitude(datum *location.Latitude) *location.Latitude {
	if datum == nil {
		return nil
	}
	clone := location.NewLatitude()
	clone.Units = test.CloneString(datum.Units)
	clone.Value = test.CloneFloat64(datum.Value)
	return clone
}
