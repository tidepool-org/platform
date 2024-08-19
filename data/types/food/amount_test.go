package food_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	dataTypesFoodTest "github.com/tidepool-org/platform/data/types/food/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Amount", func() {
	It("AmountUnitsLengthMaximum is expected", func() {
		Expect(dataTypesFood.AmountUnitsLengthMaximum).To(Equal(100))
	})

	It("AmountValueMinimum is expected", func() {
		Expect(dataTypesFood.AmountValueMinimum).To(Equal(0.0))
	})

	Context("Amount", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesFood.Amount)) {
				datum := dataTypesFoodTest.RandomAmount()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesFoodTest.NewObjectFromAmount(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesFoodTest.NewObjectFromAmount(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesFood.Amount) {},
			),
			Entry("empty",
				func(datum *dataTypesFood.Amount) {
					*datum = *dataTypesFood.NewAmount()
				},
			),
			Entry("all",
				func(datum *dataTypesFood.Amount) {
					datum.Units = pointer.FromString(test.RandomStringFromRange(1, dataTypesFood.AmountUnitsLengthMaximum))
					datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesFood.AmountValueMinimum, test.RandomFloat64Maximum()))
				},
			),
		)

		Context("ParseAmount", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesFood.ParseAmount(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesFoodTest.RandomAmount()
				object := dataTypesFoodTest.NewObjectFromAmount(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesFood.ParseAmount(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewAmount", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesFood.NewAmount()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Units).To(BeNil())
				Expect(datum.Value).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesFood.Amount), expectedErrors ...error) {
					expectedDatum := dataTypesFoodTest.RandomAmount()
					object := dataTypesFoodTest.NewObjectFromAmount(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesFood.NewAmount()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesFood.Amount) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesFood.Amount) {
						object["units"] = true
						object["value"] = true
						expectedDatum.Units = nil
						expectedDatum.Value = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/units"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/value"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesFood.Amount), expectedErrors ...error) {
					datum := dataTypesFoodTest.RandomAmount()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesFood.Amount) {},
				),
				Entry("units missing",
					func(datum *dataTypesFood.Amount) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units empty",
					func(datum *dataTypesFood.Amount) { datum.Units = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/units"),
				),
				Entry("units invalid",
					func(datum *dataTypesFood.Amount) {
						datum.Units = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/units"),
				),
				Entry("units valid",
					func(datum *dataTypesFood.Amount) {
						datum.Units = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
				),
				Entry("value missing",
					func(datum *dataTypesFood.Amount) { datum.Value = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("value out of range (lower)",
					func(datum *dataTypesFood.Amount) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-0.1, 0.0), "/value"),
				),
				Entry("value in range (lower)",
					func(datum *dataTypesFood.Amount) { datum.Value = pointer.FromFloat64(0.0) },
				),
				Entry("value in range (upper)",
					func(datum *dataTypesFood.Amount) { datum.Value = pointer.FromFloat64(math.MaxFloat64) },
				),
				Entry("multiple errors",
					func(datum *dataTypesFood.Amount) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
			)
		})
	})
})
