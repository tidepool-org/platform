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
)

var _ = Describe("RecommendedBasal", func() {
	Context("ParseRecommendedBasal", func() {
		// TODO
	})

	Context("NewRecommendedBasal", func() {
		It("is successful", func() {
			Expect(dataTypesDosingDecision.NewRecommendedBasal()).To(Equal(&dataTypesDosingDecision.RecommendedBasal{}))
		})
	})

	Context("RecommendedBasal", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesDosingDecision.RecommendedBasal), expectedErrors ...error) {
					datum := dataTypesDosingDecisionTest.RandomRecommendedBasal()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesDosingDecision.RecommendedBasal) {},
				),
				Entry("rate missing",
					func(datum *dataTypesDosingDecision.RecommendedBasal) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("rate below minimum",
					func(datum *dataTypesDosingDecision.RecommendedBasal) {
						datum.Rate = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 100), "/rate"),
				),
				Entry("rate above maximum",
					func(datum *dataTypesDosingDecision.RecommendedBasal) {
						datum.Rate = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0, 100), "/rate"),
				),
				Entry("duration missing",
					func(datum *dataTypesDosingDecision.RecommendedBasal) { datum.Duration = nil },
				),
				Entry("duration below minimum",
					func(datum *dataTypesDosingDecision.RecommendedBasal) {
						datum.Duration = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration"),
				),
				Entry("duration above maximum",
					func(datum *dataTypesDosingDecision.RecommendedBasal) {
						datum.Duration = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration"),
				),
				Entry("multiple errors",
					func(datum *dataTypesDosingDecision.RecommendedBasal) {
						datum.Rate = nil
						datum.Duration = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration"),
				),
			)
		})
	})
})
