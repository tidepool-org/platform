package status_test

import (
	"github.com/tidepool-org/platform/data/types/devicestatus/status"
	"github.com/tidepool-org/platform/pointer"
)

func NewReservoirRemaining() *status.ReservoirRemaining {
	datum := *status.NewReservoirRemaining()
	datum.Unit = pointer.FromString("mls")
	datum.Amount = pointer.FromFloat64(20.0)
	return &datum
}
