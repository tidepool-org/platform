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

func RandomBolusState() *pumpstatus.BolusState {
	datum := *pumpstatus.NewBolusState()
	datum.State = pointer.FromString(test.RandomStringFromArray(pumpstatus.BolusStates()))
	datum.DoseEntry = data.RandomDoseEntry()
	datum.Date = pointer.FromString(test.FutureNearTime().Format(time.RFC3339Nano))

	return &datum
}

var _ = Describe("BolusState", func() {

	Context("BolusState", func() {
		Context("Validate", func() {
			DescribeTable("return the expected results when the input",

				func(mutator func(datum *pumpstatus.BolusState), expectedErrors ...error) {
					datum := RandomBolusState()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pumpstatus.BolusState) {},
				),
				Entry("State does not exists",
					func(datum *pumpstatus.BolusState) {
						datum.State = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
				),
				Entry("State invalid",
					func(datum *pumpstatus.BolusState) {
						datum.State = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", pumpstatus.BolusStates()), "/state"),
				),
				Entry("Date does not exists",
					func(datum *pumpstatus.BolusState) {
						datum.Date = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/date"),
				),
				Entry("Date invalid",
					func(datum *pumpstatus.BolusState) {
						datum.Date = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/date"),
				),
				Entry("No Dose Entry Structure",
					func(datum *pumpstatus.BolusState) {
						datum.DoseEntry = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/doseEntry"),
				),
				Entry("Multiple Errors",
					func(datum *pumpstatus.BolusState) {
						datum.State = pointer.FromString("invalid")
						datum.Date = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", pumpstatus.BolusStates()), "/state"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/date"),
				),
			)
		})
	})
})
