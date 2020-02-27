package pumpstatus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	dataTest "github.com/tidepool-org/platform/data/test"

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
	if *datum.State == pumpstatus.InProgress {
		datum.DoseEntry = dataTest.RandomDoseEntry()
	} else {
		datum.DoseEntry = nil
	}
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
						datum.DoseEntry = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", pumpstatus.BolusStates()), "/state"),
				),
				Entry("No Dose Entry Structure for in progress",
					func(datum *pumpstatus.BolusState) {
						datum.State = pointer.FromString(pumpstatus.InProgress)
						datum.DoseEntry = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/doseEntry"),
				),
				Entry("Dose Entry Structure for Initiating",
					func(datum *pumpstatus.BolusState) {
						datum.State = pointer.FromString(pumpstatus.Initiating)
						datum.DoseEntry = dataTest.RandomDoseEntry()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/doseEntry"),
				),
			)
		})
	})
})
