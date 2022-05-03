package food_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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

var _ = Describe("Protein", func() {
	It("ProteinTotalGramsMaximum is expected", func() {
		Expect(dataTypesFood.ProteinTotalGramsMaximum).To(Equal(1000.0))
	})

	It("ProteinTotalGramsMinimum is expected", func() {
		Expect(dataTypesFood.ProteinTotalGramsMinimum).To(Equal(0.0))
	})

	It("ProteinUnitsGrams is expected", func() {
		Expect(dataTypesFood.ProteinUnitsGrams).To(Equal("grams"))
	})

	It("ProteinUnits returns expected", func() {
		Expect(dataTypesFood.ProteinUnits()).To(Equal([]string{"grams"}))
	})

	Context("Protein", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesFood.Protein)) {
				datum := dataTypesFoodTest.RandomProtein()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesFoodTest.NewObjectFromProtein(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesFoodTest.NewObjectFromProtein(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesFood.Protein) {},
			),
			Entry("empty",
				func(datum *dataTypesFood.Protein) {
					*datum = *dataTypesFood.NewProtein()
				},
			),
			Entry("all",
				func(datum *dataTypesFood.Protein) {
					datum.Total = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesFood.ProteinTotalGramsMinimum, dataTypesFood.ProteinTotalGramsMaximum))
					datum.Units = pointer.FromString(test.RandomStringFromArray(dataTypesFood.ProteinUnits()))
				},
			),
		)

		Context("ParseProtein", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesFood.ParseProtein(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesFoodTest.RandomProtein()
				object := dataTypesFoodTest.NewObjectFromProtein(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesFood.ParseProtein(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewProtein", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesFood.NewProtein()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Total).To(BeNil())
				Expect(datum.Units).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesFood.Protein), expectedErrors ...error) {
					expectedDatum := dataTypesFoodTest.RandomProtein()
					object := dataTypesFoodTest.NewObjectFromProtein(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesFood.NewProtein()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesFood.Protein) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesFood.Protein) {
						object["total"] = true
						object["units"] = true
						expectedDatum.Total = nil
						expectedDatum.Units = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/total"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/units"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesFood.Protein), expectedErrors ...error) {
					datum := dataTypesFoodTest.RandomProtein()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesFood.Protein) {},
				),
				Entry("total missing",
					func(datum *dataTypesFood.Protein) { datum.Total = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/total"),
				),
				Entry("total out of range (lower)",
					func(datum *dataTypesFood.Protein) { datum.Total = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/total"),
				),
				Entry("total in range (lower)",
					func(datum *dataTypesFood.Protein) { datum.Total = pointer.FromFloat64(0.0) },
				),
				Entry("total in range (upper)",
					func(datum *dataTypesFood.Protein) { datum.Total = pointer.FromFloat64(1000.0) },
				),
				Entry("total out of range (upper)",
					func(datum *dataTypesFood.Protein) { datum.Total = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/total"),
				),
				Entry("units missing",
					func(datum *dataTypesFood.Protein) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *dataTypesFood.Protein) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"grams"}), "/units"),
				),
				Entry("units grams",
					func(datum *dataTypesFood.Protein) { datum.Units = pointer.FromString("grams") },
				),
				Entry("multiple errors",
					func(datum *dataTypesFood.Protein) {
						datum.Total = nil
						datum.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/total"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})
	})
})
