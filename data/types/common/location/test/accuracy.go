package test

import (
	"github.com/tidepool-org/platform/data/types/common/location"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewAccuracy(units *string) *location.Accuracy {
	datum := location.NewAccuracy()
	datum.Units = units
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(location.AccuracyValueRangeForUnits(units)))
	return datum
}

func CloneAccuracy(datum *location.Accuracy) *location.Accuracy {
	if datum == nil {
		return nil
	}
	clone := location.NewAccuracy()
	clone.Units = test.CloneString(datum.Units)
	clone.Value = test.CloneFloat64(datum.Value)
	return clone
}
