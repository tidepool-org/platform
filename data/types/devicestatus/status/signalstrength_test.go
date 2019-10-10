package status_test

import (
	"github.com/tidepool-org/platform/data/types/devicestatus/status"
	"github.com/tidepool-org/platform/pointer"
)

func NewSignalStrength() *status.SignalStrength {
	datum := *status.NewSignalStrength()
	datum.Unit = pointer.FromString("ounces")
	datum.Value = pointer.FromFloat64(10.0)
	return &datum
}
