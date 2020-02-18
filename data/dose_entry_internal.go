package data

import (
	"time"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomDoseEntry() *DoseEntry {
	d := NewDoseEntry()
	d.StartDate = pointer.FromString(test.FutureNearTime().Format(time.RFC3339Nano))
	d.EndDate = pointer.FromString(test.FutureFarTime().Format(time.RFC3339Nano))

	d.DoseType = pointer.FromString(test.RandomStringFromArray(DoseTypes()))
	d.Unit = pointer.FromString(test.RandomStringFromArray(DoseUnits()))
	d.Value = pointer.FromFloat64(test.RandomFloat64FromRange(MinValue, MaxValue))
	d.DeliveredUnits = pointer.FromFloat64(test.RandomFloat64FromRange(MinDeliveredUnits, MaxDeliveredUnits))
	d.Description = pointer.FromString("Description")
	d.SyncIdentifier = pointer.FromString("SyncIdentifier")
	d.ScheduledBasalRate = pointer.FromFloat64(test.RandomFloat64FromRange(MinBasalRate, MaxBasalRate))

	return d
}
