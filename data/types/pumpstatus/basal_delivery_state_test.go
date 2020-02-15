package pumpstatus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"time"

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
	datum.DoseEntry = data.RandomDoseEntry()
	datum.Date = pointer.FromString(test.FutureNearTime().Format(time.RFC3339Nano))

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
				Entry("State invalid",
					func(datum *pumpstatus.BasalDeliveryState) {
						datum.State = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", pumpstatus.BasalDeliveryStates()), "/state"),
				),
				Entry("Date invalid",
					func(datum *pumpstatus.BasalDeliveryState) {
						datum.Date = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/date"),
				),
				Entry("Multiple Errors",
					func(datum *pumpstatus.BasalDeliveryState) {
						datum.State = pointer.FromString("invalid")
						datum.Date = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", pumpstatus.BasalDeliveryStates()), "/state"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/date"),
				),
			)
		})
	})
})
