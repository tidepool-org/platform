package status_test

import (
	"github.com/tidepool-org/platform/data/types/devicestatus/status"
	"github.com/tidepool-org/platform/pointer"
)

func NewBattery() *status.Battery {
	datum := *status.NewBattery()
	datum.Unit = pointer.FromString("grams")
	datum.Value = pointer.FromFloat64(5.0)
	return &datum
}
