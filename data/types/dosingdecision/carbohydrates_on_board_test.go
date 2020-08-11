package dosingdecision_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

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
				Entry("endTime before startTime",
					func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) {
						datum.StartTime = pointer.FromTime(test.PastNearTime())
						datum.EndTime = pointer.FromTime(test.PastFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.PastNearTime()), "/endTime"),
				),
				Entry("amount missing",
					func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) { datum.Amount = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("amount; out of range (lower)",
					func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) {
						datum.Amount = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 1000), "/amount"),
				),
				Entry("amount; in range (lower)",
					func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) {
						datum.Amount = pointer.FromFloat64(0)
					},
				),
				Entry("amount; in range (upper)",
					func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) {
						datum.Amount = pointer.FromFloat64(1000)
					},
				),
				Entry("amount; out of range (upper)",
					func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) {
						datum.Amount = pointer.FromFloat64(1000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0, 1000), "/amount"),
				),
				Entry("multiple errors",
					func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) {
						datum.StartTime = pointer.FromTime(test.PastNearTime())
						datum.EndTime = pointer.FromTime(test.PastFarTime())
						datum.Amount = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.PastNearTime()), "/endTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
			)
		})
	})
})
