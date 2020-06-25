package test

import (
	"github.com/tidepool-org/platform/data/types/bolus/biphasic"
	"github.com/tidepool-org/platform/data/types/bolus/normal"
	dataTypesCommonTest "github.com/tidepool-org/platform/data/types/common/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewLinkedBolus() *biphasic.LinkedBolus {
	datum := biphasic.NewLinkedBolus()
	datum.Normal = pointer.FromFloat64(test.RandomFloat64FromRange(normal.NormalMinimum, normal.NormalMaximum))
	datum.Duration = dataTypesCommonTest.NewDuration()
	return datum
}

func CloneLinkedBolus(datum *biphasic.LinkedBolus) *biphasic.LinkedBolus {
	if datum == nil {
		return nil
	}
	clone := biphasic.NewLinkedBolus()
	clone.Normal = pointer.CloneFloat64(datum.Normal)
	clone.Duration = dataTypesCommonTest.CloneDuration(datum.Duration)
	return clone
}
