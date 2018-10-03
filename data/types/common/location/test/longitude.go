package test

import (
	"github.com/tidepool-org/platform/data/types/common/location"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewLongitude() *location.Longitude {
	datum := location.NewLongitude()
	datum.Units = pointer.FromString("degrees")
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(-180.0, 180.0))
	return datum
}

func CloneLongitude(datum *location.Longitude) *location.Longitude {
	if datum == nil {
		return nil
	}
	clone := location.NewLongitude()
	clone.Units = test.CloneString(datum.Units)
	clone.Value = test.CloneFloat64(datum.Value)
	return clone
}
