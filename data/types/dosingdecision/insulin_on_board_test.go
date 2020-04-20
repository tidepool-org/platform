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
)

var _ = Describe("InsulinOnBoard", func() {
	Context("ParseInsulinOnBoard", func() {
		// TODO
	})

	Context("NewInsulinOnBoard", func() {
		It("is successful", func() {
			Expect(dataTypesDosingDecision.NewInsulinOnBoard()).To(Equal(&dataTypesDosingDecision.InsulinOnBoard{}))
		})
	})

	Context("InsulinOnBoard", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesDosingDecision.InsulinOnBoard), expectedErrors ...error) {
					datum := dataTypesDosingDecisionTest.RandomInsulinOnBoard()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesDosingDecision.InsulinOnBoard) {},
				),
				Entry("startTime invalid",
					func(datum *dataTypesDosingDecision.InsulinOnBoard) { datum.StartTime = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/startTime"),
				),
				Entry("amount missing",
					func(datum *dataTypesDosingDecision.InsulinOnBoard) { datum.Amount = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("amount; out of range (lower)",
					func(datum *dataTypesDosingDecision.InsulinOnBoard) {
						datum.Amount = pointer.FromFloat64(-1000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1000.1, -1000, 1000), "/amount"),
				),
				Entry("amount; in range (lower)",
					func(datum *dataTypesDosingDecision.InsulinOnBoard) {
						datum.Amount = pointer.FromFloat64(-1000)
					},
				),
				Entry("amount; in range (upper)",
					func(datum *dataTypesDosingDecision.InsulinOnBoard) {
						datum.Amount = pointer.FromFloat64(1000)
					},
				),
				Entry("amount; out of range (upper)",
					func(datum *dataTypesDosingDecision.InsulinOnBoard) {
						datum.Amount = pointer.FromFloat64(1000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, -1000, 1000), "/amount"),
				),
				Entry("multiple errors",
					func(datum *dataTypesDosingDecision.InsulinOnBoard) {
						datum.StartTime = pointer.FromString("invalid")
						datum.Amount = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/startTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
			)
		})
	})
})
