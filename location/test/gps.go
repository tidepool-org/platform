package test

import (
	"github.com/tidepool-org/platform/location"
	originTest "github.com/tidepool-org/platform/origin/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomGPS() *location.GPS {
	datum := location.NewGPS()
	datum.Elevation = RandomElevation()
	datum.Floor = pointer.FromInt(RandomFloor())
	datum.HorizontalAccuracy = RandomAccuracy()
	datum.Latitude = RandomLatitude()
	datum.Longitude = RandomLongitude()
	datum.Origin = originTest.RandomOrigin()
	datum.VerticalAccuracy = RandomAccuracy()
	return datum
}

func CloneGPS(datum *location.GPS) *location.GPS {
	if datum == nil {
		return nil
	}
	clone := location.NewGPS()
	clone.Elevation = CloneElevation(datum.Elevation)
	clone.Floor = pointer.CloneInt(datum.Floor)
	clone.HorizontalAccuracy = CloneAccuracy(datum.HorizontalAccuracy)
	clone.Latitude = CloneLatitude(datum.Latitude)
	clone.Longitude = CloneLongitude(datum.Longitude)
	clone.Origin = originTest.CloneOrigin(datum.Origin)
	clone.VerticalAccuracy = CloneAccuracy(datum.VerticalAccuracy)
	return clone
}

func NewObjectFromGPS(datum *location.GPS, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Elevation != nil {
		object["elevation"] = NewObjectFromElevation(datum.Elevation, objectFormat)
	}
	if datum.Floor != nil {
		object["floor"] = test.NewObjectFromInt(*datum.Floor, objectFormat)
	}
	if datum.HorizontalAccuracy != nil {
		object["horizontalAccuracy"] = NewObjectFromAccuracy(datum.HorizontalAccuracy, objectFormat)
	}
	if datum.Latitude != nil {
		object["latitude"] = NewObjectFromLatitude(datum.Latitude, objectFormat)
	}
	if datum.Longitude != nil {
		object["longitude"] = NewObjectFromLongitude(datum.Longitude, objectFormat)
	}
	if datum.Origin != nil {
		object["origin"] = originTest.NewObjectFromOrigin(datum.Origin, objectFormat)
	}
	if datum.VerticalAccuracy != nil {
		object["verticalAccuracy"] = NewObjectFromAccuracy(datum.VerticalAccuracy, objectFormat)
	}
	return object
}

func RandomFloor() int {
	return test.RandomIntFromRange(location.GPSFloorMinimum, location.GPSFloorMaximum)
}
