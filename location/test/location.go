package test

import (
	"github.com/tidepool-org/platform/location"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomLocation() *location.Location {
	datum := location.NewLocation()
	datum.GPS = RandomGPS()
	datum.Name = pointer.FromString(RandomName())
	return datum
}

func CloneLocation(datum *location.Location) *location.Location {
	if datum == nil {
		return nil
	}
	clone := location.NewLocation()
	clone.GPS = CloneGPS(datum.GPS)
	clone.Name = pointer.CloneString(datum.Name)
	return clone
}

func NewObjectFromLocation(datum *location.Location, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.GPS != nil {
		object["gps"] = NewObjectFromGPS(datum.GPS, objectFormat)
	}
	if datum.Name != nil {
		object["name"] = test.NewObjectFromString(*datum.Name, objectFormat)
	}
	return object
}

func RandomName() string {
	return test.RandomStringFromRange(1, location.NameLengthMaximum)
}
