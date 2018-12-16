package test

import (
	"github.com/tidepool-org/platform/data/types/common/location"
	dataTypesCommonOriginTest "github.com/tidepool-org/platform/data/types/common/origin/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewGPS() *location.GPS {
	datum := location.NewGPS()
	datum.Elevation = NewElevation(pointer.FromString("feet"))
	datum.Floor = pointer.FromInt(test.RandomIntFromRange(-1000, 1000))
	datum.HorizontalAccuracy = NewAccuracy(pointer.FromString("feet"))
	datum.Latitude = NewLatitude()
	datum.Longitude = NewLongitude()
	datum.Origin = dataTypesCommonOriginTest.NewOrigin()
	datum.VerticalAccuracy = NewAccuracy(pointer.FromString("feet"))
	return datum
}

func CloneGPS(datum *location.GPS) *location.GPS {
	if datum == nil {
		return nil
	}
	clone := location.NewGPS()
	clone.Elevation = CloneElevation(datum.Elevation)
	clone.Floor = test.CloneInt(datum.Floor)
	clone.HorizontalAccuracy = CloneAccuracy(datum.HorizontalAccuracy)
	clone.Latitude = CloneLatitude(datum.Latitude)
	clone.Longitude = CloneLongitude(datum.Longitude)
	clone.Origin = dataTypesCommonOriginTest.CloneOrigin(datum.Origin)
	clone.VerticalAccuracy = CloneAccuracy(datum.VerticalAccuracy)
	return clone
}
