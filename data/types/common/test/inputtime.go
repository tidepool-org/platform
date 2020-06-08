package test

import (
	"time"

	dataTypesCommon "github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/test"

	"github.com/tidepool-org/platform/pointer"
)

func NewInputTime() *dataTypesCommon.InputTime {
	datum := dataTypesCommon.NewInputTime()
	timeReference := test.RandomTime()
	datum.InputTime = pointer.FromString(timeReference.Format(time.RFC3339Nano))
	return datum
}

func CloneInputTime(datum *dataTypesCommon.InputTime) *dataTypesCommon.InputTime {
	if datum == nil {
		return nil
	}
	clone := dataTypesCommon.NewInputTime()
	clone.InputTime = pointer.CloneString(datum.InputTime)
	return clone
}
