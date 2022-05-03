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

var _ = Describe("Ingredient", func() {
	It("IngredientArrayLengthMaximum is expected", func() {
		Expect(dataTypesFood.IngredientArrayLengthMaximum).To(Equal(100))
	})

	It("IngredientBrandLengthMaximum is expected", func() {
		Expect(dataTypesFood.IngredientBrandLengthMaximum).To(Equal(100))
	})

	It("IngredientCodeLengthMaximum is expected", func() {
		Expect(dataTypesFood.IngredientCodeLengthMaximum).To(Equal(100))
	})

	It("IngredientNameLengthMaximum is expected", func() {
		Expect(dataTypesFood.IngredientNameLengthMaximum).To(Equal(100))
	})

	Context("Ingredient", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesFood.Ingredient)) {
				datum := dataTypesFoodTest.RandomIngredient(3)
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesFoodTest.NewObjectFromIngredient(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesFoodTest.NewObjectFromIngredient(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesFood.Ingredient) {},
			),
			Entry("empty",
				func(datum *dataTypesFood.Ingredient) {
					*datum = *dataTypesFood.NewIngredient()
				},
			),
			Entry("all",
				func(datum *dataTypesFood.Ingredient) {
					datum.Amount = dataTypesFoodTest.RandomAmount()
					datum.Brand = pointer.FromString(test.RandomStringFromRange(1, dataTypesFood.IngredientBrandLengthMaximum))
					datum.Code = pointer.FromString(test.RandomStringFromRange(1, dataTypesFood.IngredientCodeLengthMaximum))
					datum.Ingredients = dataTypesFoodTest.RandomIngredientArray(3)
					datum.Name = pointer.FromString(test.RandomStringFromRange(1, dataTypesFood.IngredientNameLengthMaximum))
					datum.Nutrition = dataTypesFoodTest.RandomNutrition()
				},
			),
		)

		Context("ParseIngredient", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesFood.ParseIngredient(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesFoodTest.RandomIngredient(3)
				object := dataTypesFoodTest.NewObjectFromIngredient(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesFood.ParseIngredient(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewIngredient", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesFood.NewIngredient()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Amount).To(BeNil())
				Expect(datum.Brand).To(BeNil())
				Expect(datum.Code).To(BeNil())
				Expect(datum.Ingredients).To(BeNil())
				Expect(datum.Name).To(BeNil())
				Expect(datum.Nutrition).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesFood.Ingredient), expectedErrors ...error) {
					expectedDatum := dataTypesFoodTest.RandomIngredient(3)
					object := dataTypesFoodTest.NewObjectFromIngredient(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesFood.NewIngredient()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesFood.Ingredient) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesFood.Ingredient) {
						object["amount"] = true
						object["brand"] = true
						object["code"] = true
						object["ingredients"] = true
						object["name"] = true
						object["nutrition"] = true
						expectedDatum.Amount = nil
						expectedDatum.Brand = nil
						expectedDatum.Code = nil
						expectedDatum.Ingredients = nil
						expectedDatum.Name = nil
						expectedDatum.Nutrition = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/amount"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/brand"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/code"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/ingredients"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/name"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/nutrition"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesFood.Ingredient), expectedErrors ...error) {
					datum := dataTypesFoodTest.RandomIngredient(3)
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesFood.Ingredient) {},
				),
				Entry("amount missing",
					func(datum *dataTypesFood.Ingredient) { datum.Amount = nil },
				),
				Entry("amount invalid",
					func(datum *dataTypesFood.Ingredient) { datum.Amount.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount/units"),
				),
				Entry("amount valid",
					func(datum *dataTypesFood.Ingredient) { datum.Amount = dataTypesFoodTest.RandomAmount() },
				),
				Entry("brand missing",
					func(datum *dataTypesFood.Ingredient) { datum.Brand = nil },
				),
				Entry("brand empty",
					func(datum *dataTypesFood.Ingredient) { datum.Brand = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/brand"),
				),
				Entry("brand length; in range (upper)",
					func(datum *dataTypesFood.Ingredient) {
						datum.Brand = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("brand length; out of range (upper)",
					func(datum *dataTypesFood.Ingredient) {
						datum.Brand = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/brand"),
				),
				Entry("code missing",
					func(datum *dataTypesFood.Ingredient) { datum.Code = nil },
				),
				Entry("code empty",
					func(datum *dataTypesFood.Ingredient) { datum.Code = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/code"),
				),
				Entry("code length; in range (upper)",
					func(datum *dataTypesFood.Ingredient) {
						datum.Code = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("code length; out of range (upper)",
					func(datum *dataTypesFood.Ingredient) {
						datum.Code = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/code"),
				),
				Entry("ingredients missing",
					func(datum *dataTypesFood.Ingredient) { datum.Ingredients = nil },
				),
				Entry("ingredients invalid",
					func(datum *dataTypesFood.Ingredient) { (*datum.Ingredients)[0] = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/ingredients/0"),
				),
				Entry("ingredients valid",
					func(datum *dataTypesFood.Ingredient) { datum.Ingredients = dataTypesFoodTest.RandomIngredientArray(3) },
				),
				Entry("name missing",
					func(datum *dataTypesFood.Ingredient) { datum.Name = nil },
				),
				Entry("name empty",
					func(datum *dataTypesFood.Ingredient) { datum.Name = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
				Entry("name length; in range (upper)",
					func(datum *dataTypesFood.Ingredient) {
						datum.Name = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("name length; out of range (upper)",
					func(datum *dataTypesFood.Ingredient) {
						datum.Name = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/name"),
				),
				Entry("nutrition missing",
					func(datum *dataTypesFood.Ingredient) { datum.Nutrition = nil },
				),
				Entry("nutrition invalid",
					func(datum *dataTypesFood.Ingredient) { datum.Nutrition.Carbohydrate.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/nutrition/carbohydrate/units"),
				),
				Entry("nutrition valid",
					func(datum *dataTypesFood.Ingredient) { datum.Nutrition = dataTypesFoodTest.RandomNutrition() },
				),
				Entry("multiple errors",
					func(datum *dataTypesFood.Ingredient) {
						datum.Amount.Units = nil
						datum.Brand = pointer.FromString("")
						datum.Code = pointer.FromString("")
						(*datum.Ingredients)[0] = nil
						datum.Name = pointer.FromString("")
						datum.Nutrition.Carbohydrate.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/brand"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/code"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/ingredients/0"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/nutrition/carbohydrate/units"),
				),
			)
		})
	})

	Context("IngredientArray", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesFood.IngredientArray)) {
				datum := dataTypesFoodTest.RandomIngredientArray(3)
				mutator(datum)
				test.ExpectSerializedArrayJSON(dataTypesFoodTest.AnonymizeIngredientArray(datum), dataTypesFoodTest.NewArrayFromIngredientArray(datum, test.ObjectFormatJSON))
				test.ExpectSerializedArrayBSON(dataTypesFoodTest.AnonymizeIngredientArray(datum), dataTypesFoodTest.NewArrayFromIngredientArray(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesFood.IngredientArray) {},
			),
			Entry("empty",
				func(datum *dataTypesFood.IngredientArray) {
					*datum = *dataTypesFood.NewIngredientArray()
				},
			),
		)

		Context("ParseIngredientArray", func() {
			It("returns nil when the array is missing", func() {
				Expect(dataTypesFood.ParseIngredientArray(structureParser.NewArray(nil))).To(BeNil())
			})

			It("returns new datum when the array is valid", func() {
				datum := dataTypesFoodTest.RandomIngredientArray(3)
				array := dataTypesFoodTest.NewArrayFromIngredientArray(datum, test.ObjectFormatJSON)
				parser := structureParser.NewArray(&array)
				Expect(dataTypesFood.ParseIngredientArray(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewIngredientArray", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesFood.NewIngredientArray()
				Expect(datum).ToNot(BeNil())
				Expect(*datum).To(BeEmpty())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object []interface{}, expectedDatum *dataTypesFood.IngredientArray), expectedErrors ...error) {
					expectedDatum := dataTypesFoodTest.RandomIngredientArray(3)
					array := dataTypesFoodTest.NewArrayFromIngredientArray(expectedDatum, test.ObjectFormatJSON)
					mutator(array, expectedDatum)
					datum := dataTypesFood.NewIngredientArray()
					errorsTest.ExpectEqual(structureParser.NewArray(&array).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object []interface{}, expectedDatum *dataTypesFood.IngredientArray) {},
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesFood.IngredientArray), expectedErrors ...error) {
					datum := dataTypesFood.NewIngredientArray()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesFood.IngredientArray) {},
					structureValidator.ErrorValueEmpty(),
				),
				Entry("empty",
					func(datum *dataTypesFood.IngredientArray) { *datum = *dataTypesFood.NewIngredientArray() },
					structureValidator.ErrorValueEmpty(),
				),
				Entry("nil",
					func(datum *dataTypesFood.IngredientArray) { *datum = append(*datum, nil) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					func(datum *dataTypesFood.IngredientArray) {
						invalid := dataTypesFoodTest.RandomIngredient(3)
						invalid.Brand = pointer.FromString("")
						*datum = append(*datum, invalid)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/0/brand"),
				),
				Entry("single valid",
					func(datum *dataTypesFood.IngredientArray) {
						*datum = append(*datum, dataTypesFoodTest.RandomIngredient(3))
					},
				),
				Entry("multiple invalid",
					func(datum *dataTypesFood.IngredientArray) {
						invalid := dataTypesFoodTest.RandomIngredient(3)
						invalid.Brand = pointer.FromString("")
						*datum = append(*datum, dataTypesFoodTest.RandomIngredient(3), invalid, dataTypesFoodTest.RandomIngredient(3))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/1/brand"),
				),
				Entry("multiple valid",
					func(datum *dataTypesFood.IngredientArray) {
						*datum = append(*datum, dataTypesFoodTest.RandomIngredient(3), dataTypesFoodTest.RandomIngredient(3), dataTypesFoodTest.RandomIngredient(3))
					},
				),
				Entry("multiple; length in range (upper)",
					func(datum *dataTypesFood.IngredientArray) {
						for len(*datum) < 100 {
							*datum = append(*datum, dataTypesFoodTest.RandomIngredient(1))
						}
					},
				),
				Entry("multiple; length out of range (upper)",
					func(datum *dataTypesFood.IngredientArray) {
						for len(*datum) < 101 {
							*datum = append(*datum, dataTypesFoodTest.RandomIngredient(1))
						}
					},
					structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100),
				),
				Entry("multiple errors",
					func(datum *dataTypesFood.IngredientArray) {
						invalid := dataTypesFoodTest.RandomIngredient(3)
						invalid.Brand = pointer.FromString("")
						*datum = append(*datum, nil, invalid, dataTypesFoodTest.RandomIngredient(3))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/1/brand"),
				),
			)
		})
	})
})
