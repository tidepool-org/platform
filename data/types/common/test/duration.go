package commontypes_test

import (
	commontypes "github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewDuration() *commontypes.Duration {
	datum := commontypes.NewDuration()
	datum.Units = pointer.FromString(test.RandomStringFromArray(commontypes.DurationUnits()))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(commontypes.DurationValueRangeForUnits(datum.Units)))
	return datum
}

func CloneDuration(datum *commontypes.Duration) *commontypes.Duration {
	if datum == nil {
		return nil
	}
	clone := commontypes.NewDuration()
	clone.Units = pointer.CloneString(datum.Units)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}
