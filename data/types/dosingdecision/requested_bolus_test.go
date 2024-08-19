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

var _ = Describe("RequestedBolus", func() {
	It("RequestedBolusAmountMaximum is expected", func() {
		Expect(dataTypesDosingDecision.RequestedBolusAmountMaximum).To(Equal(1000))
	})

	It("RequestedBolusAmountMinimum is expected", func() {
		Expect(dataTypesDosingDecision.RequestedBolusAmountMinimum).To(Equal(0))
	})

	Context("RequestedBolus", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesDosingDecision.RequestedBolus)) {
				datum := dataTypesDosingDecisionTest.RandomRequestedBolus()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesDosingDecisionTest.NewObjectFromRequestedBolus(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesDosingDecisionTest.NewObjectFromRequestedBolus(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesDosingDecision.RequestedBolus) {},
			),
			Entry("empty",
				func(datum *dataTypesDosingDecision.RequestedBolus) {
					*datum = *dataTypesDosingDecision.NewRequestedBolus()
				},
			),
			Entry("all",
				func(datum *dataTypesDosingDecision.RequestedBolus) {
					datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingDecision.RequestedBolusAmountMinimum, dataTypesDosingDecision.RequestedBolusAmountMaximum))
				},
			),
		)

		Context("ParseRequestedBolus", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesDosingDecision.ParseRequestedBolus(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesDosingDecisionTest.RandomRequestedBolus()
				object := dataTypesDosingDecisionTest.NewObjectFromRequestedBolus(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesDosingDecision.ParseRequestedBolus(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewRequestedBolus", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesDosingDecision.NewRequestedBolus()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Amount).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.RequestedBolus), expectedErrors ...error) {
					expectedDatum := dataTypesDosingDecisionTest.RandomRequestedBolus()
					object := dataTypesDosingDecisionTest.NewObjectFromRequestedBolus(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesDosingDecision.NewRequestedBolus()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.RequestedBolus) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.RequestedBolus) {
						object["amount"] = true
						expectedDatum.Amount = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/amount"),
				),
			)
		})
		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesDosingDecision.RequestedBolus), expectedErrors ...error) {
					datum := dataTypesDosingDecisionTest.RandomRequestedBolus()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesDosingDecision.RequestedBolus) {},
				),
				Entry("amount missing",
					func(datum *dataTypesDosingDecision.RequestedBolus) { datum.Amount = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("amount; out of range (lower)",
					func(datum *dataTypesDosingDecision.RequestedBolus) {
						datum.Amount = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 1000), "/amount"),
				),
				Entry("amount; in range (lower)",
					func(datum *dataTypesDosingDecision.RequestedBolus) {
						datum.Amount = pointer.FromFloat64(0)
					},
				),
				Entry("amount; in range (upper)",
					func(datum *dataTypesDosingDecision.RequestedBolus) {
						datum.Amount = pointer.FromFloat64(1000)
					},
				),
				Entry("amount; out of range (upper)",
					func(datum *dataTypesDosingDecision.RequestedBolus) {
						datum.Amount = pointer.FromFloat64(1000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0, 1000), "/amount"),
				),
			)
		})
	})
})
