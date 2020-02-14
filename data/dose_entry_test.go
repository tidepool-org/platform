package data_test

import (
	"time"

	"github.com/tidepool-org/platform/structure"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/pointer"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/test"

	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
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

var _ = Describe("DoseEntry", func() {
	Context("DoseEntry", func() {
		Context("Validate", func() {
			DescribeTable("return the expected results when the input",

				func(mutator func(datum *data.DoseEntry), expectedErrors ...error) {
					datum := RandomDoseEntry()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *data.DoseEntry) {},
				),
				Entry("Dose type invalid",
					func(datum *data.DoseEntry) {
						datum.DoseType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", data.DoseTypes()), "/doseType"),
				),
				Entry("Unit type invalid",
					func(datum *data.DoseEntry) {
						datum.Unit = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", data.DoseUnits()), "/unit"),
				),
				Entry("StartDate invalid",
					func(datum *data.DoseEntry) {
						datum.StartDate = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/startDate"),
				),
				Entry("EndDate invalid",
					func(datum *data.DoseEntry) {
						datum.EndDate = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/endDate"),
				),
				Entry("end date not after start date",
					func(datum *data.DoseEntry) {
						datum.EndDate = pointer.FromString(test.PastFarTime().Format(time.RFC3339Nano))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.FutureNearTime()), "/endDate"),
				),
				Entry("Value below Minimum",
					func(datum *data.DoseEntry) {
						datum.Value = pointer.FromFloat64(data.MinValue - 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(data.MinValue-1, data.MinValue, data.MaxValue), "/value"),
				),
				Entry("Value above Maximum",
					func(datum *data.DoseEntry) {
						datum.Value = pointer.FromFloat64(data.MaxValue + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(data.MaxValue+1, data.MinValue, data.MaxValue), "/value"),
				),
				Entry("DeliveredUnits below Minimum",
					func(datum *data.DoseEntry) {
						datum.DeliveredUnits = pointer.FromFloat64(data.MinDeliveredUnits - 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(data.MinDeliveredUnits-1, data.MinDeliveredUnits, data.MaxDeliveredUnits), "/deliveredUnits"),
				),
				Entry("DeliveredUnits above Maximum",
					func(datum *data.DoseEntry) {
						datum.DeliveredUnits = pointer.FromFloat64(data.MaxDeliveredUnits + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(data.MaxDeliveredUnits+1, data.MinDeliveredUnits, data.MaxDeliveredUnits), "/deliveredUnits"),
				),
				Entry("ScheduledBasalRate below Minimum",
					func(datum *data.DoseEntry) {
						datum.ScheduledBasalRate = pointer.FromFloat64(data.MinBasalRate - 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(data.MinBasalRate-1, data.MinBasalRate, data.MaxBasalRate), "/scheduledBasalRate"),
				),
				Entry("ScheduledBasalRate above Maximum",
					func(datum *data.DoseEntry) {
						datum.ScheduledBasalRate = pointer.FromFloat64(data.MaxBasalRate + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(data.MaxBasalRate+1, data.MinBasalRate, data.MaxBasalRate), "/scheduledBasalRate"),
				),
				Entry("Multiple Errors",
					func(datum *data.DoseEntry) {
						datum.DoseType = pointer.FromString("invalid")
						datum.EndDate = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(data.MinValue - 1)
						datum.DeliveredUnits = pointer.FromFloat64(data.MaxDeliveredUnits + 1)
						datum.ScheduledBasalRate = pointer.FromFloat64(data.MaxBasalRate + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", data.DoseTypes()), "/doseType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/endDate"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(data.MinValue-1, data.MinValue, data.MaxValue), "/value"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(data.MaxDeliveredUnits+1, data.MinDeliveredUnits, data.MaxDeliveredUnits), "/deliveredUnits"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(data.MaxBasalRate+1, data.MinBasalRate, data.MaxBasalRate), "/scheduledBasalRate"),
				),
			)
		})
	})
})
