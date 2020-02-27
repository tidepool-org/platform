package test

import (
	"time"

	"github.com/tidepool-org/platform/data"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomDoseEntry() *data.DoseEntry {
	d := data.NewDoseEntry()
	d.StartDate = pointer.FromString(test.FutureNearTime().Format(time.RFC3339Nano))
	d.EndDate = pointer.FromString(test.FutureFarTime().Format(time.RFC3339Nano))

	d.DoseType = pointer.FromString(test.RandomStringFromArray(data.DoseTypes()))
	d.Unit = pointer.FromString(test.RandomStringFromArray(data.DoseUnits()))
	d.Value = pointer.FromFloat64(test.RandomFloat64FromRange(data.MinValue, data.MaxValue))
	d.DeliveredUnits = pointer.FromFloat64(test.RandomFloat64FromRange(data.MinDeliveredUnits, data.MaxDeliveredUnits))
	d.Description = pointer.FromString("Description")
	d.SyncIdentifier = pointer.FromString("SyncIdentifier")
	d.ScheduledBasalRate = pointer.FromFloat64(test.RandomFloat64FromRange(data.MinBasalRate, data.MaxBasalRate))

	return d
}
