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

var _ = Describe("InsulinOnBoard", func() {
	It("InsulinOnBoardAmountMaximum is expected", func() {
		Expect(dataTypesDosingDecision.InsulinOnBoardAmountMaximum).To(Equal(1000))
	})

	It("InsulinOnBoardAmountMinimum is expected", func() {
		Expect(dataTypesDosingDecision.InsulinOnBoardAmountMinimum).To(Equal(-1000))
	})

	Context("InsulinOnBoard", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesDosingDecision.InsulinOnBoard)) {
				datum := dataTypesDosingDecisionTest.RandomInsulinOnBoard()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesDosingDecisionTest.NewObjectFromInsulinOnBoard(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesDosingDecisionTest.NewObjectFromInsulinOnBoard(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesDosingDecision.InsulinOnBoard) {},
			),
			Entry("empty",
				func(datum *dataTypesDosingDecision.InsulinOnBoard) {
					*datum = *dataTypesDosingDecision.NewInsulinOnBoard()
				},
			),
			Entry("all",
				func(datum *dataTypesDosingDecision.InsulinOnBoard) {
					datum.Time = pointer.FromTime(test.RandomTime())
					datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingDecision.InsulinOnBoardAmountMinimum, dataTypesDosingDecision.InsulinOnBoardAmountMaximum))
				},
			),
		)

		Context("ParseInsulinOnBoard", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesDosingDecision.ParseInsulinOnBoard(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesDosingDecisionTest.RandomInsulinOnBoard()
				object := dataTypesDosingDecisionTest.NewObjectFromInsulinOnBoard(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesDosingDecision.ParseInsulinOnBoard(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewInsulinOnBoard", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesDosingDecision.NewInsulinOnBoard()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Time).To(BeNil())
				Expect(datum.Amount).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.InsulinOnBoard), expectedErrors ...error) {
					expectedDatum := dataTypesDosingDecisionTest.RandomInsulinOnBoard()
					object := dataTypesDosingDecisionTest.NewObjectFromInsulinOnBoard(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesDosingDecision.NewInsulinOnBoard()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.InsulinOnBoard) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.InsulinOnBoard) {
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
				func(mutator func(datum *dataTypesDosingDecision.InsulinOnBoard), expectedErrors ...error) {
					datum := dataTypesDosingDecisionTest.RandomInsulinOnBoard()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesDosingDecision.InsulinOnBoard) {},
				),
				Entry("time missing",
					func(datum *dataTypesDosingDecision.InsulinOnBoard) {
						datum.Time = nil
					},
				),
				Entry("time exists",
					func(datum *dataTypesDosingDecision.InsulinOnBoard) {
						datum.Time = pointer.FromTime(test.RandomTime())
					},
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
						datum.Amount = pointer.FromFloat64(0)
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
						datum.Amount = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
			)
		})
	})
})
