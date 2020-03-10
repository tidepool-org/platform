package dosingdecision_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"time"

	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	dataTypesDosingDecisionTest "github.com/tidepool-org/platform/data/types/dosingdecision/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("CarbohydratesOnBoard", func() {
	Context("ParseCarbohydratesOnBoard", func() {
		// TODO
	})

	Context("NewCarbohydratesOnBoard", func() {
		It("is successful", func() {
			Expect(dataTypesDosingDecision.NewCarbohydratesOnBoard()).To(Equal(&dataTypesDosingDecision.CarbohydratesOnBoard{}))
		})
	})

	Context("CarbohydratesOnBoard", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesDosingDecision.CarbohydratesOnBoard), expectedErrors ...error) {
					datum := dataTypesDosingDecisionTest.RandomCarbohydratesOnBoard()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) {},
				),
				Entry("startTime invalid",
					func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) {
						datum.StartTime = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/startTime"),
				),
				Entry("endTime invalid",
					func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) {
						datum.EndTime = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/endTime"),
				),
				Entry("endTime before startTime",
					func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) {
						datum.StartTime = pointer.FromString(test.PastNearTime().Format(dataTypesDosingDecision.TimeFormat))
						datum.EndTime = pointer.FromString(test.PastFarTime().Format(dataTypesDosingDecision.TimeFormat))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.PastNearTime()), "/endTime"),
				),
				Entry("amount missing",
					func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) { datum.Amount = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("amount below minimum",
					func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) {
						datum.Amount = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 1000), "/amount"),
				),
				Entry("amount above maximum",
					func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) {
						datum.Amount = pointer.FromFloat64(1000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0, 1000), "/amount"),
				),
				Entry("multiple errors",
					func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) {
						datum.StartTime = pointer.FromString("invalid")
						datum.EndTime = pointer.FromString("invalid")
						datum.Amount = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/startTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/endTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
			)
		})
	})
})
