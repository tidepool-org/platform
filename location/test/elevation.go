package test

import (
	"github.com/tidepool-org/platform/location"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomElevation() *location.Elevation {
	datum := location.NewElevation()
	datum.Units = pointer.FromString(RandomElevationUnits())
	datum.Value = pointer.FromFloat64(RandomElevationValue(datum.Units))
	return datum
}

func CloneElevation(datum *location.Elevation) *location.Elevation {
	if datum == nil {
		return nil
	}
	clone := location.NewElevation()
	clone.Units = pointer.CloneString(datum.Units)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}

func NewObjectFromElevation(datum *location.Elevation, objectFormat test.ObjectFormat) map[string]interface{} {
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

func RandomElevationUnits() string {
	return test.RandomStringFromArray(location.ElevationUnits())
}

func RandomElevationValue(units *string) float64 {
	return test.RandomFloat64FromRange(location.ElevationValueRangeForUnits(units))
}
