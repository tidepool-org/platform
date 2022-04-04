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
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("RecommendedBolus", func() {
	It("RecommendedBolusAmountMaximum is expected", func() {
		Expect(dataTypesDosingDecision.RecommendedBolusAmountMaximum).To(Equal(1000))
	})

	It("RecommendedBolusAmountMinimum is expected", func() {
		Expect(dataTypesDosingDecision.RecommendedBolusAmountMinimum).To(Equal(0))
	})

	Context("RecommendedBolus", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesDosingDecision.RecommendedBolus)) {
				datum := dataTypesDosingDecisionTest.RandomRecommendedBolus()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesDosingDecisionTest.NewObjectFromRecommendedBolus(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesDosingDecisionTest.NewObjectFromRecommendedBolus(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesDosingDecision.RecommendedBolus) {},
			),
			Entry("empty",
				func(datum *dataTypesDosingDecision.RecommendedBolus) {
					*datum = *dataTypesDosingDecision.NewRecommendedBolus()
				},
			),
			Entry("all",
				func(datum *dataTypesDosingDecision.RecommendedBolus) {
					datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingDecision.RecommendedBolusAmountMinimum, dataTypesDosingDecision.RecommendedBolusAmountMaximum))
				},
			),
		)

		Context("ParseRecommendedBolus", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesDosingDecision.ParseRecommendedBolus(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesDosingDecisionTest.RandomRecommendedBolus()
				object := dataTypesDosingDecisionTest.NewObjectFromRecommendedBolus(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesDosingDecision.ParseRecommendedBolus(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewRecommendedBolus", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesDosingDecision.NewRecommendedBolus()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Amount).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.RecommendedBolus), expectedErrors ...error) {
					expectedDatum := dataTypesDosingDecisionTest.RandomRecommendedBolus()
					object := dataTypesDosingDecisionTest.NewObjectFromRecommendedBolus(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesDosingDecision.NewRecommendedBolus()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.RecommendedBolus) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.RecommendedBolus) {
						object["amount"] = true
						expectedDatum.Amount = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/amount"),
				),
			)
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
				Entry("amount missing",
					func(datum *dataTypesDosingDecision.RecommendedBolus) { datum.Amount = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("amount; out of range (lower)",
					func(datum *dataTypesDosingDecision.RecommendedBolus) {
						datum.Amount = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 1000), "/amount"),
				),
				Entry("amount; in range (lower)",
					func(datum *dataTypesDosingDecision.RecommendedBolus) {
						datum.Amount = pointer.FromFloat64(0)
					},
				),
				Entry("amount; in range (upper)",
					func(datum *dataTypesDosingDecision.RecommendedBolus) {
						datum.Amount = pointer.FromFloat64(1000)
					},
				),
				Entry("amount; out of range (upper)",
					func(datum *dataTypesDosingDecision.RecommendedBolus) {
						datum.Amount = pointer.FromFloat64(1000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0, 1000), "/amount"),
				),
			)
		})
	})
})
