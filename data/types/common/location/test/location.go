package test

import (
	"github.com/tidepool-org/platform/data/types/common/location"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewLocation() *location.Location {
	datum := location.NewLocation()
	datum.GPS = NewGPS()
	datum.Name = pointer.String(test.NewText(1, 100))
	return datum
}

func CloneLocation(datum *location.Location) *location.Location {
	if datum == nil {
		return nil
	}
	clone := location.NewLocation()
	clone.GPS = CloneGPS(datum.GPS)
	clone.Name = test.CloneString(datum.Name)
	return clone
}
