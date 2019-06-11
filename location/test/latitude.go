package test

import (
	"github.com/tidepool-org/platform/location"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomLatitude() *location.Latitude {
	datum := location.NewLatitude()
	datum.Units = pointer.FromString(RandomLatitudeUnits())
	datum.Value = pointer.FromFloat64(RandomLatitudeValue(datum.Units))
	return datum
}

func CloneLatitude(datum *location.Latitude) *location.Latitude {
	if datum == nil {
		return nil
	}
	clone := location.NewLatitude()
	clone.Units = pointer.CloneString(datum.Units)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}

func NewObjectFromLatitude(datum *location.Latitude, objectFormat test.ObjectFormat) map[string]interface{} {
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

func RandomLatitudeUnits() string {
	return test.RandomStringFromArray(location.LatitudeUnits())
}

func RandomLatitudeValue(units *string) float64 {
	return test.RandomFloat64FromRange(location.LatitudeValueRangeForUnits(units))
}
