package pumpstatus_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"

	"github.com/tidepool-org/platform/data/types/pumpstatus"

	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func RandomBasalDeliveryState() *pumpstatus.BasalDeliveryState {
	datum := *pumpstatus.NewBasalDeliveryState()
	datum.State = pointer.FromString(test.RandomStringFromArray(pumpstatus.BasalDeliveryStates()))
	if *datum.State == pumpstatus.Active || *datum.State == pumpstatus.Suspended {
		datum.Date = pointer.FromString(test.FutureNearTime().Format(time.RFC3339Nano))
	} else {
		datum.Date = nil
	}
	if *datum.State == pumpstatus.TempBasal {
		datum.DoseEntry = data.RandomDoseEntry()
	} else {
		datum.DoseEntry = nil
	}

	return &datum
}

var _ = Describe("BasalDeliveryState", func() {

	Context("BasalDeliveryState", func() {
		Context("Validate", func() {
			DescribeTable("return the expected results when the input",

				func(mutator func(datum *pumpstatus.BasalDeliveryState), expectedErrors ...error) {
					datum := RandomBasalDeliveryState()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pumpstatus.BasalDeliveryState) {},
				),
				Entry("State does not exists",
					func(datum *pumpstatus.BasalDeliveryState) {
						datum.State = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
				),
				Entry("State invalid",
					func(datum *pumpstatus.BasalDeliveryState) {
						datum.State = pointer.FromString("invalid")
						datum.DoseEntry = nil
						datum.Date = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", pumpstatus.BasalDeliveryStates()), "/state"),
				),
				Entry("Date does not exists",
					func(datum *pumpstatus.BasalDeliveryState) {
						datum.State = pointer.FromString(pumpstatus.Active)
						datum.Date = nil
						datum.DoseEntry = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/date"),
				),
				Entry("Date invalid",
					func(datum *pumpstatus.BasalDeliveryState) {
						datum.State = pointer.FromString(pumpstatus.Active)
						datum.Date = pointer.FromString("invalid")
						datum.DoseEntry = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/date"),
				),
				Entry("No Dose Entry Structure",
					func(datum *pumpstatus.BasalDeliveryState) {
						datum.State = pointer.FromString(pumpstatus.TempBasal)
						datum.Date = nil
						datum.DoseEntry = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/doseEntry"),
				),
				Entry("Dose Entry Structure on wrong state",
					func(datum *pumpstatus.BasalDeliveryState) {
						datum.State = pointer.FromString(pumpstatus.Active)
						datum.DoseEntry = data.RandomDoseEntry()
						datum.Date = pointer.FromString(test.FutureNearTime().Format(time.RFC3339Nano))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/doseEntry"),
				),
				Entry("Date on wrong state",
					func(datum *pumpstatus.BasalDeliveryState) {
						datum.State = pointer.FromString(pumpstatus.TempBasal)
						datum.DoseEntry = data.RandomDoseEntry()
						datum.Date = pointer.FromString(test.FutureNearTime().Format(time.RFC3339Nano))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/date"),
				),
				Entry("Multiple Errors",
					func(datum *pumpstatus.BasalDeliveryState) {
						datum.State = pointer.FromString(pumpstatus.Active)
						datum.Date = pointer.FromString("invalid")
						datum.DoseEntry = data.RandomDoseEntry()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/date"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/doseEntry"),
				),
			)
		})
	})
})
