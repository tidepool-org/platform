package test

import (
	"github.com/tidepool-org/platform/location"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomLongitude() *location.Longitude {
	datum := location.NewLongitude()
	datum.Units = pointer.FromString(RandomLongitudeUnits())
	datum.Value = pointer.FromFloat64(RandomLongitudeValue(datum.Units))
	return datum
}

func CloneLongitude(datum *location.Longitude) *location.Longitude {
	if datum == nil {
		return nil
	}
	clone := location.NewLongitude()
	clone.Units = pointer.CloneString(datum.Units)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}

func NewObjectFromLongitude(datum *location.Longitude, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Units != nil {
		object["units"] = test.NewObjectFromString(*datum.Units, objectFormat)
	}
	if datum.Value != nil {
		object["value"] = test.NewObjectFromFloat64(*datum.Value, objectFormat)
	}
	return object
}

func RandomLongitudeUnits() string {
	return test.RandomStringFromArray(location.LongitudeUnits())
}

func RandomLongitudeValue(units *string) float64 {
	return test.RandomFloat64FromRange(location.LongitudeValueRangeForUnits(units))
}
