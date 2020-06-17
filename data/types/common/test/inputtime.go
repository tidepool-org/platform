package test

import (
	"time"

	"github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/test"

	"github.com/tidepool-org/platform/pointer"
)

func NewInputTime() *common.InputTime {
	datum := common.NewInputTime()
	timeReference := test.RandomTime()
	datum.InputTime = pointer.FromString(timeReference.Format(time.RFC3339Nano))
	return datum
}

func CloneInputTime(datum *common.InputTime) *common.InputTime {
	if datum == nil {
		return nil
	}
	clone := common.NewInputTime()
	clone.InputTime = pointer.CloneString(datum.InputTime)
	return clone
}
