package test

import (
	dataTypesCommon "github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewDuration() *dataTypesCommon.Duration {
	datum := dataTypesCommon.NewDuration()
	datum.Units = pointer.FromString(test.RandomStringFromArray(dataTypesCommon.DurationUnits()))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesCommon.DurationValueRangeForUnits(datum.Units)))
	return datum
}

func CloneDuration(datum *dataTypesCommon.Duration) *dataTypesCommon.Duration {
	if datum == nil {
		return nil
	}
	clone := dataTypesCommon.NewDuration()
	clone.Units = pointer.CloneString(datum.Units)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}
