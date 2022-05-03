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

var _ = Describe("CarbohydratesOnBoard", func() {
	It("CarbohydratesOnBoardAmountMaximum is expected", func() {
		Expect(dataTypesDosingDecision.CarbohydratesOnBoardAmountMaximum).To(Equal(1000))
	})

	It("CarbohydratesOnBoardAmountMinimum is expected", func() {
		Expect(dataTypesDosingDecision.CarbohydratesOnBoardAmountMinimum).To(Equal(0))
	})

	Context("CarbohydratesOnBoard", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesDosingDecision.CarbohydratesOnBoard)) {
				datum := dataTypesDosingDecisionTest.RandomCarbohydratesOnBoard()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesDosingDecisionTest.NewObjectFromCarbohydratesOnBoard(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesDosingDecisionTest.NewObjectFromCarbohydratesOnBoard(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) {},
			),
			Entry("empty",
				func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) {
					*datum = *dataTypesDosingDecision.NewCarbohydratesOnBoard()
				},
			),
			Entry("all",
				func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) {
					datum.Time = pointer.FromTime(test.RandomTime())
					datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingDecision.CarbohydratesOnBoardAmountMinimum, dataTypesDosingDecision.CarbohydratesOnBoardAmountMaximum))
				},
			),
		)

		Context("ParseCarbohydratesOnBoard", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesDosingDecision.ParseCarbohydratesOnBoard(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesDosingDecisionTest.RandomCarbohydratesOnBoard()
				object := dataTypesDosingDecisionTest.NewObjectFromCarbohydratesOnBoard(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesDosingDecision.ParseCarbohydratesOnBoard(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewCarbohydratesOnBoard", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesDosingDecision.NewCarbohydratesOnBoard()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Time).To(BeNil())
				Expect(datum.Amount).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.CarbohydratesOnBoard), expectedErrors ...error) {
					expectedDatum := dataTypesDosingDecisionTest.RandomCarbohydratesOnBoard()
					object := dataTypesDosingDecisionTest.NewObjectFromCarbohydratesOnBoard(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesDosingDecision.NewCarbohydratesOnBoard()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.CarbohydratesOnBoard) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.CarbohydratesOnBoard) {
						object["time"] = true
						object["amount"] = true
						expectedDatum.Time = nil
						expectedDatum.Amount = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/time"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/amount"),
				),
			)
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
				Entry("time missing",
					func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) {
						datum.Time = nil
					},
				),
				Entry("time exists",
					func(datum *dataTypesDosingDecision.CarbohydratesOnBoard) {
						datum.Time = pointer.FromTime(test.RandomTime())
					},
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
						datum.Amount = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
			)
		})
	})
})
