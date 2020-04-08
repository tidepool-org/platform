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

var _ = Describe("RecommendedBolus", func() {
	Context("ParseRecommendedBolus", func() {
		// TODO
	})

	Context("NewRecommendedBolus", func() {
		It("is successful", func() {
			Expect(dataTypesDosingDecision.NewRecommendedBolus()).To(Equal(&dataTypesDosingDecision.RecommendedBolus{}))
		})
	})

	Context("RecommendedBolus", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesDosingDecision.RecommendedBolus), expectedErrors ...error) {
					datum := dataTypesDosingDecisionTest.RandomRecommendedBolus()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesDosingDecision.RecommendedBolus) {},
				),
				Entry("time invalid",
					func(datum *dataTypesDosingDecision.RecommendedBolus) { datum.Time = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/time"),
				),
				Entry("amount missing",
					func(datum *dataTypesDosingDecision.RecommendedBolus) { datum.Amount = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("amount below minimum",
					func(datum *dataTypesDosingDecision.RecommendedBolus) {
						datum.Amount = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 1000), "/amount"),
				),
				Entry("amount above maximum",
					func(datum *dataTypesDosingDecision.RecommendedBolus) {
						datum.Amount = pointer.FromFloat64(1000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0, 1000), "/amount"),
				),
				Entry("multiple errors",
					func(datum *dataTypesDosingDecision.RecommendedBolus) {
						datum.Time = pointer.FromString("invalid")
						datum.Amount = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/time"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
			)
		})
	})
})
