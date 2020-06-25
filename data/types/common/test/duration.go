package test

import (
	"github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewDuration() *common.Duration {
	datum := common.NewDuration()
	datum.Units = pointer.FromString(test.RandomStringFromArray(common.DurationUnits()))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(common.DurationValueRangeForUnits(datum.Units)))
	return datum
}

func CloneDuration(datum *common.Duration) *common.Duration {
	if datum == nil {
		return nil
	}
	clone := common.NewDuration()
	clone.Units = pointer.CloneString(datum.Units)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}
