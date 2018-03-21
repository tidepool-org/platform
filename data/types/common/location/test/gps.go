package test

import (
	"github.com/tidepool-org/platform/data/types/common/location"
	testDataTypesCommonOrigin "github.com/tidepool-org/platform/data/types/common/origin/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewGPS() *location.GPS {
	datum := location.NewGPS()
	datum.Elevation = NewElevation(pointer.String("feet"))
	datum.Floor = pointer.Int(test.RandomIntFromRange(-1000, 1000))
	datum.HorizontalAccuracy = NewAccuracy(pointer.String("feet"))
	datum.Latitude = NewLatitude()
	datum.Longitude = NewLongitude()
	datum.Origin = testDataTypesCommonOrigin.NewOrigin()
	datum.VerticalAccuracy = NewAccuracy(pointer.String("feet"))
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
	clone.Origin = testDataTypesCommonOrigin.CloneOrigin(datum.Origin)
	clone.VerticalAccuracy = CloneAccuracy(datum.VerticalAccuracy)
	return clone
}
