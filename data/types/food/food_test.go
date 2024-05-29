package food_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	dataTypes "github.com/tidepool-org/platform/data/types"
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

func NewMeta() interface{} {
	return &dataTypes.Meta{
		Type: "food",
	}
}

var _ = Describe("Food", func() {
	It("Type is expected", func() {
		Expect(dataTypesFood.Type).To(Equal("food"))
	})

	It("BrandLengthMaximum is expected", func() {
		Expect(dataTypesFood.BrandLengthMaximum).To(Equal(100))
	})

	It("CodeLengthMaximum is expected", func() {
		Expect(dataTypesFood.CodeLengthMaximum).To(Equal(100))
	})

	It("MealBreakfast is expected", func() {
		Expect(dataTypesFood.MealBreakfast).To(Equal("breakfast"))
	})

	It("MealDinner is expected", func() {
		Expect(dataTypesFood.MealDinner).To(Equal("dinner"))
	})

	It("MealLunch is expected", func() {
		Expect(dataTypesFood.MealLunch).To(Equal("lunch"))
	})

	It("MealOther is expected", func() {
		Expect(dataTypesFood.MealOther).To(Equal("other"))
	})

	It("MealOtherLengthMaximum is expected", func() {
		Expect(dataTypesFood.MealOtherLengthMaximum).To(Equal(100))
	})

	It("MealSnack is expected", func() {
		Expect(dataTypesFood.MealSnack).To(Equal("snack"))
	})

	It("NameLengthMaximum is expected", func() {
		Expect(dataTypesFood.NameLengthMaximum).To(Equal(100))
	})

	It("Meals returns expected", func() {
		Expect(dataTypesFood.Meals()).To(Equal([]string{"breakfast", "dinner", "lunch", "other", "snack"}))
	})

	Context("Food", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesFood.Food)) {
				datum := dataTypesFoodTest.RandomFood(3)
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesFoodTest.NewObjectFromFood(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesFoodTest.NewObjectFromFood(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesFood.Food) {},
			),
			Entry("empty",
				func(datum *dataTypesFood.Food) {
					*datum = *dataTypesFood.New()
				},
			),
			Entry("all",
				func(datum *dataTypesFood.Food) {
					datum.Amount = dataTypesFoodTest.RandomAmount()
					datum.Brand = pointer.FromString(test.RandomStringFromRange(1, dataTypesFood.BrandLengthMaximum))
					datum.Code = pointer.FromString(test.RandomStringFromRange(1, dataTypesFood.CodeLengthMaximum))
					datum.Ingredients = dataTypesFoodTest.RandomIngredientArray(3)
					datum.Meal = pointer.FromString(test.RandomStringFromArray(dataTypesFood.Meals()))
					datum.MealOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					datum.Name = pointer.FromString(test.RandomStringFromRange(1, 100))
					datum.Nutrition = dataTypesFoodTest.RandomNutrition()
				},
			),
		)

		Context("New", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesFood.New()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Type).To(Equal("food"))
				Expect(datum.Amount).To(BeNil())
				Expect(datum.Brand).To(BeNil())
				Expect(datum.Code).To(BeNil())
				Expect(datum.Ingredients).To(BeNil())
				Expect(datum.Meal).To(BeNil())
				Expect(datum.MealOther).To(BeNil())
				Expect(datum.Name).To(BeNil())
				Expect(datum.Nutrition).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesFood.Food), expectedErrors ...error) {
					expectedDatum := dataTypesFoodTest.RandomFoodForParser(3)
					object := dataTypesFoodTest.NewObjectFromFood(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesFood.New()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesFood.Food) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesFood.Food) {
						object["amount"] = true
						object["brand"] = true
						object["code"] = true
						object["ingredients"] = true
						object["meal"] = true
						object["mealOther"] = true
						object["name"] = true
						object["nutrition"] = true
						expectedDatum.Amount = nil
						expectedDatum.Brand = nil
						expectedDatum.Code = nil
						expectedDatum.Ingredients = nil
						expectedDatum.Meal = nil
						expectedDatum.MealOther = nil
						expectedDatum.Name = nil
						expectedDatum.Nutrition = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/amount", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotString(true), "/brand", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotString(true), "/code", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotArray(true), "/ingredients", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotString(true), "/meal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotString(true), "/mealOther", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotString(true), "/name", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/nutrition", NewMeta()),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesFood.Food), expectedErrors ...error) {
					datum := dataTypesFoodTest.RandomFood(3)
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesFood.Food) {},
				),
				Entry("type missing",
					func(datum *dataTypesFood.Food) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypes.Meta{}),
				),
				Entry("type invalid",
					func(datum *dataTypesFood.Food) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "food"), "/type", &dataTypes.Meta{Type: "invalidType"}),
				),
				Entry("type food",
					func(datum *dataTypesFood.Food) { datum.Type = "food" },
				),
				Entry("amount missing",
					func(datum *dataTypesFood.Food) { datum.Amount = nil },
				),
				Entry("amount invalid",
					func(datum *dataTypesFood.Food) { datum.Amount.Units = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/amount/units", NewMeta()),
				),
				Entry("amount valid",
					func(datum *dataTypesFood.Food) { datum.Amount = dataTypesFoodTest.RandomAmount() },
				),
				Entry("brand missing",
					func(datum *dataTypesFood.Food) { datum.Brand = nil },
				),
				Entry("brand empty",
					func(datum *dataTypesFood.Food) { datum.Brand = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/brand", NewMeta()),
				),
				Entry("brand length; in range (upper)",
					func(datum *dataTypesFood.Food) {
						datum.Brand = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("brand length; out of range (upper)",
					func(datum *dataTypesFood.Food) {
						datum.Brand = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/brand", NewMeta()),
				),
				Entry("code missing",
					func(datum *dataTypesFood.Food) { datum.Code = nil },
				),
				Entry("code empty",
					func(datum *dataTypesFood.Food) { datum.Code = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/code", NewMeta()),
				),
				Entry("code length; in range (upper)",
					func(datum *dataTypesFood.Food) { datum.Code = pointer.FromString(test.RandomStringFromRange(100, 100)) },
				),
				Entry("code length; out of range (upper)",
					func(datum *dataTypesFood.Food) { datum.Code = pointer.FromString(test.RandomStringFromRange(101, 101)) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/code", NewMeta()),
				),
				Entry("ingredients missing",
					func(datum *dataTypesFood.Food) { datum.Ingredients = nil },
				),
				Entry("ingredients invalid",
					func(datum *dataTypesFood.Food) { (*datum.Ingredients)[0] = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/ingredients/0", NewMeta()),
				),
				Entry("ingredients valid",
					func(datum *dataTypesFood.Food) { datum.Ingredients = dataTypesFoodTest.RandomIngredientArray(3) },
				),
				Entry("meal missing; meal other missing",
					func(datum *dataTypesFood.Food) {
						datum.Meal = nil
						datum.MealOther = nil
					},
				),
				Entry("meal missing; meal other exists",
					func(datum *dataTypesFood.Food) {
						datum.Meal = nil
						datum.MealOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/mealOther", NewMeta()),
				),
				Entry("meal invalid; meal other missing",
					func(datum *dataTypesFood.Food) {
						datum.Meal = pointer.FromString("invalid")
						datum.MealOther = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"breakfast", "dinner", "lunch", "other", "snack"}), "/meal", NewMeta()),
				),
				Entry("meal invalid; meal other exists",
					func(datum *dataTypesFood.Food) {
						datum.Meal = pointer.FromString("invalid")
						datum.MealOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"breakfast", "dinner", "lunch", "other", "snack"}), "/meal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/mealOther", NewMeta()),
				),
				Entry("meal breakfast; meal other missing",
					func(datum *dataTypesFood.Food) {
						datum.Meal = pointer.FromString("breakfast")
						datum.MealOther = nil
					},
				),
				Entry("meal breakfast; meal other exists",
					func(datum *dataTypesFood.Food) {
						datum.Meal = pointer.FromString("breakfast")
						datum.MealOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/mealOther", NewMeta()),
				),
				Entry("meal dinner; meal other missing",
					func(datum *dataTypesFood.Food) {
						datum.Meal = pointer.FromString("dinner")
						datum.MealOther = nil
					},
				),
				Entry("meal dinner; meal other exists",
					func(datum *dataTypesFood.Food) {
						datum.Meal = pointer.FromString("dinner")
						datum.MealOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/mealOther", NewMeta()),
				),
				Entry("meal lunch; meal other missing",
					func(datum *dataTypesFood.Food) {
						datum.Meal = pointer.FromString("lunch")
						datum.MealOther = nil
					},
				),
				Entry("meal lunch; meal other exists",
					func(datum *dataTypesFood.Food) {
						datum.Meal = pointer.FromString("lunch")
						datum.MealOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/mealOther", NewMeta()),
				),
				Entry("meal other; meal other missing",
					func(datum *dataTypesFood.Food) {
						datum.Meal = pointer.FromString("other")
						datum.MealOther = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/mealOther", NewMeta()),
				),
				Entry("meal other; meal other empty",
					func(datum *dataTypesFood.Food) {
						datum.Meal = pointer.FromString("other")
						datum.MealOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/mealOther", NewMeta()),
				),
				Entry("meal other; meal other length; in range (upper)",
					func(datum *dataTypesFood.Food) {
						datum.Meal = pointer.FromString("other")
						datum.MealOther = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("meal other; meal other length; out of range (upper)",
					func(datum *dataTypesFood.Food) {
						datum.Meal = pointer.FromString("other")
						datum.MealOther = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/mealOther", NewMeta()),
				),
				Entry("meal snack; meal other missing",
					func(datum *dataTypesFood.Food) {
						datum.Meal = pointer.FromString("snack")
						datum.MealOther = nil
					},
				),
				Entry("meal snack; meal other exists",
					func(datum *dataTypesFood.Food) {
						datum.Meal = pointer.FromString("snack")
						datum.MealOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/mealOther", NewMeta()),
				),
				Entry("name missing",
					func(datum *dataTypesFood.Food) { datum.Name = nil },
				),
				Entry("name empty",
					func(datum *dataTypesFood.Food) { datum.Name = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/name", NewMeta()),
				),
				Entry("name length; in range (upper)",
					func(datum *dataTypesFood.Food) { datum.Name = pointer.FromString(test.RandomStringFromRange(100, 100)) },
				),
				Entry("name length; out of range (upper)",
					func(datum *dataTypesFood.Food) { datum.Name = pointer.FromString(test.RandomStringFromRange(101, 101)) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/name", NewMeta()),
				),
				Entry("nutrition missing",
					func(datum *dataTypesFood.Food) { datum.Nutrition = nil },
				),
				Entry("nutrition invalid",
					func(datum *dataTypesFood.Food) { datum.Nutrition.Carbohydrate.Units = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/nutrition/carbohydrate/units", NewMeta()),
				),
				Entry("nutrition valid",
					func(datum *dataTypesFood.Food) { datum.Nutrition = dataTypesFoodTest.RandomNutrition() },
				),
				Entry("multiple errors",
					func(datum *dataTypesFood.Food) {
						datum.Type = "invalidType"
						datum.Amount.Units = nil
						datum.Brand = pointer.FromString("")
						datum.Code = pointer.FromString("")
						(*datum.Ingredients)[0] = nil
						datum.Meal = pointer.FromString("invalid")
						datum.MealOther = pointer.FromString(test.RandomStringFromRange(1, 100))
						datum.Name = pointer.FromString("")
						datum.Nutrition.Carbohydrate.Units = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "food"), "/type", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/amount/units", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/brand", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/code", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/ingredients/0", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"breakfast", "dinner", "lunch", "other", "snack"}), "/meal", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/mealOther", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/name", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/nutrition/carbohydrate/units", &dataTypes.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *dataTypesFood.Food)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesFoodTest.RandomFood(3)
						mutator(datum)
						expectedDatum := dataTypesFoodTest.CloneFood(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *dataTypesFood.Food) {},
				),
				Entry("does not modify the datum; amount missing",
					func(datum *dataTypesFood.Food) { datum.Amount = nil },
				),
				Entry("does not modify the datum; brand missing",
					func(datum *dataTypesFood.Food) { datum.Brand = nil },
				),
				Entry("does not modify the datum; code missing",
					func(datum *dataTypesFood.Food) { datum.Code = nil },
				),
				Entry("does not modify the datum; ingredients missing",
					func(datum *dataTypesFood.Food) { datum.Ingredients = nil },
				),
				Entry("does not modify the datum; meal missing",
					func(datum *dataTypesFood.Food) { datum.Meal = nil },
				),
				Entry("does not modify the datum; meal other missing",
					func(datum *dataTypesFood.Food) { datum.MealOther = nil },
				),
				Entry("does not modify the datum; name missing",
					func(datum *dataTypesFood.Food) { datum.Name = nil },
				),
				Entry("does not modify the datum; nutrition missing",
					func(datum *dataTypesFood.Food) { datum.Nutrition = nil },
				),
			)
		})

		Context("LegacyIdentityFields", func() {
			It("returns the expected legacy identity fields", func() {
				datum := dataTypesFoodTest.RandomFood(3)
				legacyIdentityFields, err := datum.LegacyIdentityFields()
				Expect(err).ToNot(HaveOccurred())
				Expect(legacyIdentityFields).To(Equal([]string{datum.Type, *datum.DeviceID, (*datum.Time).Format(types.LegacyFieldTimeFormat)}))
			})
		})
	})
})
