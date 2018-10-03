package test

import (
	"github.com/tidepool-org/platform/data/types/common/location"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewElevation(units *string) *location.Elevation {
	datum := location.NewElevation()
	datum.Units = units
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(location.ElevationValueRangeForUnits(units)))
	return datum
}

func CloneElevation(datum *location.Elevation) *location.Elevation {
	if datum == nil {
		return nil
	}
	clone := location.NewElevation()
	clone.Units = test.CloneString(datum.Units)
	clone.Value = test.CloneFloat64(datum.Value)
	return clone
}
