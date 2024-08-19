package dosingdecision_test

import (
	. "github.com/onsi/ginkgo/v2"
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

var _ = Describe("RecommendedBasal", func() {
	It("RecommendedBasalDurationMaximum is expected", func() {
		Expect(dataTypesDosingDecision.RecommendedBasalDurationMaximum).To(Equal(86400000))
	})

	It("RecommendedBasalDurationMinimum is expected", func() {
		Expect(dataTypesDosingDecision.RecommendedBasalDurationMinimum).To(Equal(0))
	})

	It("RecommendedBasalRateMaximum is expected", func() {
		Expect(dataTypesDosingDecision.RecommendedBasalRateMaximum).To(Equal(100))
	})

	It("RecommendedBasalRateMinimum is expected", func() {
		Expect(dataTypesDosingDecision.RecommendedBasalRateMinimum).To(Equal(0))
	})

	Context("RecommendedBasal", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesDosingDecision.RecommendedBasal)) {
				datum := dataTypesDosingDecisionTest.RandomRecommendedBasal()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesDosingDecisionTest.NewObjectFromRecommendedBasal(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesDosingDecisionTest.NewObjectFromRecommendedBasal(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesDosingDecision.RecommendedBasal) {},
			),
			Entry("empty",
				func(datum *dataTypesDosingDecision.RecommendedBasal) {
					*datum = *dataTypesDosingDecision.NewRecommendedBasal()
				},
			),
			Entry("all",
				func(datum *dataTypesDosingDecision.RecommendedBasal) {
					datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingDecision.RecommendedBasalRateMinimum, dataTypesDosingDecision.RecommendedBasalRateMaximum))
					datum.Duration = pointer.FromInt(test.RandomIntFromRange(dataTypesDosingDecision.RecommendedBasalDurationMinimum, dataTypesDosingDecision.RecommendedBasalDurationMaximum))
				},
			),
		)

		Context("ParseRecommendedBasal", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesDosingDecision.ParseRecommendedBasal(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesDosingDecisionTest.RandomRecommendedBasal()
				object := dataTypesDosingDecisionTest.NewObjectFromRecommendedBasal(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesDosingDecision.ParseRecommendedBasal(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewRecommendedBasal", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesDosingDecision.NewRecommendedBasal()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Rate).To(BeNil())
				Expect(datum.Duration).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.RecommendedBasal), expectedErrors ...error) {
					expectedDatum := dataTypesDosingDecisionTest.RandomRecommendedBasal()
					object := dataTypesDosingDecisionTest.NewObjectFromRecommendedBasal(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesDosingDecision.NewRecommendedBasal()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.RecommendedBasal) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.RecommendedBasal) {
						object["rate"] = true
						object["duration"] = true
						expectedDatum.Rate = nil
						expectedDatum.Duration = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/rate"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/duration"),
				),
			)
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
				Entry("rate; out of range (lower)",
					func(datum *dataTypesDosingDecision.RecommendedBasal) {
						datum.Rate = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 100), "/rate"),
				),
				Entry("rate; in range (lower)",
					func(datum *dataTypesDosingDecision.RecommendedBasal) {
						datum.Rate = pointer.FromFloat64(0)
					},
				),
				Entry("rate; in range (upper)",
					func(datum *dataTypesDosingDecision.RecommendedBasal) {
						datum.Rate = pointer.FromFloat64(100)
					},
				),
				Entry("rate; out of range (upper)",
					func(datum *dataTypesDosingDecision.RecommendedBasal) {
						datum.Rate = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0, 100), "/rate"),
				),
				Entry("duration missing",
					func(datum *dataTypesDosingDecision.RecommendedBasal) { datum.Duration = nil },
				),
				Entry("duration; out of range (lower)",
					func(datum *dataTypesDosingDecision.RecommendedBasal) {
						datum.Duration = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration"),
				),
				Entry("duration; in range (lower)",
					func(datum *dataTypesDosingDecision.RecommendedBasal) {
						datum.Duration = pointer.FromInt(0)
					},
				),
				Entry("duration; in range (upper)",
					func(datum *dataTypesDosingDecision.RecommendedBasal) {
						datum.Duration = pointer.FromInt(86400000)
					},
				),
				Entry("duration; out of range (upper)",
					func(datum *dataTypesDosingDecision.RecommendedBasal) {
						datum.Duration = pointer.FromInt(86400001)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration"),
				),
				Entry("multiple errors",
					func(datum *dataTypesDosingDecision.RecommendedBasal) {
						datum.Rate = nil
						datum.Duration = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration"),
				),
			)
		})
	})
})
