package data_test

import (
	"time"

	. "github.com/onsi/ginkgo"

	"github.com/tidepool-org/platform/pointer"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/test"
)

func RandomDoseEntry() *data.DoseEntry {
	startDate := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now().Add(-30*24*time.Hour))
	endDate := startDate.Add(-30 * 24 * time.Hour)

	d := data.NewDoseEntry()
	d.StartDate = pointer.FromString(startDate.Format(time.RFC3339Nano))
	d.EndDate = pointer.FromString(endDate.Format(time.RFC3339Nano))

	d.DoseType = pointer.FromString(test.RandomStringFromArray(data.DoseTypes()))
	d.Value = pointer.FromFloat64(test.RandomFloat64FromRange(data.MinValue, data.MaxValue))
	d.DeliveredUnits = pointer.FromFloat64(test.RandomFloat64FromRange(data.MinDeliveredUnits, data.MaxDeliveredUnits))
	d.Description = pointer.FromString("Description")
	d.SyncIdentifier = pointer.FromString("SyncIdentifier")
	d.ScheduledBasalRate = pointer.FromFloat64(test.RandomFloat64FromRange(data.MinBasalRate, data.MaxBasalRate))

	return d
}

var _ = Describe("Forecast", func() {
	Context("Forecast", func() {
	})

})
