package food_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/common"
	dataTypesCommonTest "github.com/tidepool-org/platform/data/types/common/test"
	"github.com/tidepool-org/platform/data/types/food"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "food",
	}
}

func NewFood(ingredientArrayDepthLimit int) *food.Food {
	datum := food.New()
	datum.Base = *dataTypesTest.NewBase()
	datum.Type = "food"
	datum.Amount = NewAmount()
	datum.Brand = pointer.FromString(test.RandomStringFromRange(1, 100))
	datum.Code = pointer.FromString(test.RandomStringFromRange(1, 100))
	datum.Ingredients = NewIngredientArray(ingredientArrayDepthLimit)
	datum.Meal = pointer.FromString(test.RandomStringFromArray(food.Meals()))
	if datum.Meal != nil && *datum.Meal == food.MealOther {
		datum.MealOther = pointer.FromString(test.RandomStringFromRange(1, 100))
	}
	datum.Name = pointer.FromString(test.RandomStringFromRange(1, 100))
	datum.Nutrition = NewNutrition()
	datum.Prescriptor = dataTypesCommonTest.NewPrescriptor()
	datum.PrescribedNutrition = NewNutrition()
	return datum
}

func CloneFood(datum *food.Food) *food.Food {
	if datum == nil {
		return nil
	}
	clone := food.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.Amount = CloneAmount(datum.Amount)
	clone.Brand = pointer.CloneString(datum.Brand)
	clone.Code = pointer.CloneString(datum.Code)
	clone.Ingredients = CloneIngredientArray(datum.Ingredients)
	clone.Meal = pointer.CloneString(datum.Meal)
	clone.MealOther = pointer.CloneString(datum.MealOther)
	clone.Name = pointer.CloneString(datum.Name)
	clone.Nutrition = CloneNutrition(datum.Nutrition)
	clone.Prescriptor = dataTypesCommonTest.ClonePrescriptor(datum.Prescriptor)
	clone.PrescribedNutrition = CloneNutrition(datum.PrescribedNutrition)
	return clone
}

var _ = Describe("Food", func() {
	It("Type is expected", func() {
		Expect(food.Type).To(Equal("food"))
	})

	It("BrandLengthMaximum is expected", func() {
		Expect(food.BrandLengthMaximum).To(Equal(100))
	})

	It("CodeLengthMaximum is expected", func() {
		Expect(food.CodeLengthMaximum).To(Equal(100))
	})

	It("MealBreakfast is expected", func() {
		Expect(food.MealBreakfast).To(Equal("breakfast"))
	})

	It("MealDinner is expected", func() {
		Expect(food.MealDinner).To(Equal("dinner"))
	})

	It("MealLunch is expected", func() {
		Expect(food.MealLunch).To(Equal("lunch"))
	})

	It("MealOther is expected", func() {
		Expect(food.MealOther).To(Equal("other"))
	})

	It("MealOtherLengthMaximum is expected", func() {
		Expect(food.MealOtherLengthMaximum).To(Equal(100))
	})

	It("MealSnack is expected", func() {
		Expect(food.MealSnack).To(Equal("snack"))
	})

	It("NameLengthMaximum is expected", func() {
		Expect(food.NameLengthMaximum).To(Equal(100))
	})

	It("Meals returns expected", func() {
		Expect(food.Meals()).To(Equal([]string{"breakfast", "dinner", "lunch", "other", "snack", "rescuecarbs"}))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := food.New()
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
			Expect(datum.Prescriptor).To(Equal(&common.Prescriptor{}))
			Expect(datum.PrescribedNutrition).To(BeNil())
		})
	})

	Context("Food", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *food.Food), expectedErrors ...error) {
					datum := NewFood(3)
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *food.Food) {},
				),
				Entry("type missing",
					func(datum *food.Food) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				),
				Entry("type invalid",
					func(datum *food.Food) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "food"), "/type", &types.Meta{Type: "invalidType"}),
				),
				Entry("type food",
					func(datum *food.Food) { datum.Type = "food" },
				),
				Entry("amount missing",
					func(datum *food.Food) { datum.Amount = nil },
				),
				Entry("amount invalid",
					func(datum *food.Food) { datum.Amount.Units = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/amount/units", NewMeta()),
				),
				Entry("amount valid",
					func(datum *food.Food) { datum.Amount = NewAmount() },
				),
				Entry("brand missing",
					func(datum *food.Food) { datum.Brand = nil },
				),
				Entry("brand empty",
					func(datum *food.Food) { datum.Brand = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/brand", NewMeta()),
				),
				Entry("brand length; in range (upper)",
					func(datum *food.Food) { datum.Brand = pointer.FromString(test.RandomStringFromRange(100, 100)) },
				),
				Entry("brand length; out of range (upper)",
					func(datum *food.Food) { datum.Brand = pointer.FromString(test.RandomStringFromRange(101, 101)) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/brand", NewMeta()),
				),
				Entry("code missing",
					func(datum *food.Food) { datum.Code = nil },
				),
				Entry("code empty",
					func(datum *food.Food) { datum.Code = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/code", NewMeta()),
				),
				Entry("code length; in range (upper)",
					func(datum *food.Food) { datum.Code = pointer.FromString(test.RandomStringFromRange(100, 100)) },
				),
				Entry("code length; out of range (upper)",
					func(datum *food.Food) { datum.Code = pointer.FromString(test.RandomStringFromRange(101, 101)) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/code", NewMeta()),
				),
				Entry("ingredients missing",
					func(datum *food.Food) { datum.Ingredients = nil },
				),
				Entry("ingredients invalid",
					func(datum *food.Food) { (*datum.Ingredients)[0] = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/ingredients/0", NewMeta()),
				),
				Entry("ingredients valid",
					func(datum *food.Food) { datum.Ingredients = NewIngredientArray(3) },
				),
				Entry("meal missing; meal other missing",
					func(datum *food.Food) {
						datum.Meal = nil
						datum.MealOther = nil
					},
				),
				Entry("meal missing; meal other exists",
					func(datum *food.Food) {
						datum.Meal = nil
						datum.MealOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/mealOther", NewMeta()),
				),
				Entry("meal invalid; meal other missing",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("invalid")
						datum.MealOther = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"breakfast", "dinner", "lunch", "other", "snack", "rescuecarbs"}), "/meal", NewMeta()),
				),
				Entry("meal invalid; meal other exists",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("invalid")
						datum.MealOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"breakfast", "dinner", "lunch", "other", "snack", "rescuecarbs"}), "/meal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/mealOther", NewMeta()),
				),
				Entry("meal breakfast; meal other missing",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("breakfast")
						datum.MealOther = nil
					},
				),
				Entry("meal breakfast; meal other exists",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("breakfast")
						datum.MealOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/mealOther", NewMeta()),
				),
				Entry("meal dinner; meal other missing",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("dinner")
						datum.MealOther = nil
					},
				),
				Entry("meal dinner; meal other exists",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("dinner")
						datum.MealOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/mealOther", NewMeta()),
				),
				Entry("meal lunch; meal other missing",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("lunch")
						datum.MealOther = nil
					},
				),
				Entry("meal lunch; meal other exists",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("lunch")
						datum.MealOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/mealOther", NewMeta()),
				),
				Entry("meal other; meal other missing",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("other")
						datum.MealOther = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/mealOther", NewMeta()),
				),
				Entry("meal other; meal other empty",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("other")
						datum.MealOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/mealOther", NewMeta()),
				),
				Entry("meal other; meal other length; in range (upper)",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("other")
						datum.MealOther = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("meal other; meal other length; out of range (upper)",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("other")
						datum.MealOther = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/mealOther", NewMeta()),
				),
				Entry("meal snack; meal other missing",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("snack")
						datum.MealOther = nil
					},
				),
				Entry("meal snack; meal other exists",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("snack")
						datum.MealOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/mealOther", NewMeta()),
				),
				Entry("meal rescuecarbs; meal other missing",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("rescuecarbs")
						datum.MealOther = nil
					},
				),
				Entry("meal rescuecarbs; meal other exists",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("rescuecarbs")
						datum.MealOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/mealOther", NewMeta()),
				),
				Entry("meal rescuecarbs; prescriptor exists",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("rescuecarbs")
						datum.MealOther = nil
						datum.Prescriptor.Prescriptor = pointer.FromString("auto")
					},
				),
				Entry("meal rescuecarbs; prescriptor and prescribedNutrition exist",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("rescuecarbs")
						datum.MealOther = nil
						datum.Prescriptor.Prescriptor = pointer.FromString("auto")
						datum.PrescribedNutrition = NewNutrition()
					},
				),
				Entry("meal rescuecarbs; invalid prescriptor ",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("rescuecarbs")
						datum.MealOther = nil
						datum.Prescriptor.Prescriptor = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"auto", "manual", "hybrid"}), "/prescriptor", NewMeta()),
				),
				Entry("meal rescuecarbs; prescriptor is hybrid and prescribedNutrition does not exist",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("rescuecarbs")
						datum.MealOther = nil
						datum.Prescriptor.Prescriptor = pointer.FromString("hybrid")
						datum.PrescribedNutrition = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/prescribedNutrition", NewMeta()),
				),
				Entry("meal rescuecarbs; prescriptor is missing and prescribedNutrition exists",
					func(datum *food.Food) {
						datum.Meal = pointer.FromString("rescuecarbs")
						datum.MealOther = nil
						datum.Prescriptor = common.NewPrescriptor()
						datum.PrescribedNutrition = NewNutrition()
					},
				),
				Entry("name missing",
					func(datum *food.Food) { datum.Name = nil },
				),
				Entry("name empty",
					func(datum *food.Food) { datum.Name = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/name", NewMeta()),
				),
				Entry("name length; in range (upper)",
					func(datum *food.Food) { datum.Name = pointer.FromString(test.RandomStringFromRange(100, 100)) },
				),
				Entry("name length; out of range (upper)",
					func(datum *food.Food) { datum.Name = pointer.FromString(test.RandomStringFromRange(101, 101)) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/name", NewMeta()),
				),
				Entry("nutrition missing",
					func(datum *food.Food) { datum.Nutrition = nil },
				),
				Entry("nutrition invalid",
					func(datum *food.Food) { datum.Nutrition.Carbohydrate.Units = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/nutrition/carbohydrate/units", NewMeta()),
				),
				Entry("nutrition valid",
					func(datum *food.Food) { datum.Nutrition = NewNutrition() },
				),
				Entry("multiple errors",
					func(datum *food.Food) {
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
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "food"), "/type", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/amount/units", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/brand", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/code", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/ingredients/0", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"breakfast", "dinner", "lunch", "other", "snack", "rescuecarbs"}), "/meal", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/mealOther", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/name", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/nutrition/carbohydrate/units", &types.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *food.Food)) {
					for _, origin := range structure.Origins() {
						datum := NewFood(3)
						datum.Meal = pointer.FromString(test.RandomStringFromArray([]string{food.MealBreakfast, food.MealDinner, food.MealLunch, food.MealOther, food.MealSnack}))
						datum.Prescriptor = nil
						datum.PrescribedNutrition = nil
						mutator(datum)
						expectedDatum := CloneFood(datum)
						expectedDatum.Prescriptor = nil
						expectedDatum.PrescribedNutrition = nil
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *food.Food) {},
				),
				Entry("does not modify the datum; amount missing",
					func(datum *food.Food) { datum.Amount = nil },
				),
				Entry("does not modify the datum; brand missing",
					func(datum *food.Food) { datum.Brand = nil },
				),
				Entry("does not modify the datum; code missing",
					func(datum *food.Food) { datum.Code = nil },
				),
				Entry("does not modify the datum; ingredients missing",
					func(datum *food.Food) { datum.Ingredients = nil },
				),
				Entry("does not modify the datum; meal missing",
					func(datum *food.Food) { datum.Meal = nil },
				),
				Entry("does not modify the datum; meal other missing",
					func(datum *food.Food) { datum.MealOther = nil },
				),
				Entry("does not modify the datum; name missing",
					func(datum *food.Food) { datum.Name = nil },
				),
				Entry("does not modify the datum; nutrition missing",
					func(datum *food.Food) { datum.Nutrition = nil },
				),
				Entry("does not crash when prescriptor is set",
					func(datum *food.Food) {
						datum.Prescriptor = dataTypesCommonTest.NewPrescriptor()
					},
				),
				Entry("does not crash when prescribedNutrition is set",
					func(datum *food.Food) {
						datum.PrescribedNutrition = NewNutrition()
					},
				),
			)

			DescribeTable("normalizes the datum for rescueCarbs, hybrid prescription",
				func(mutator func(datum *food.Food)) {
					for _, origin := range structure.Origins() {
						datum := NewFood(3)
						datum.Meal = pointer.FromString(food.MealRescueCarbs)
						datum.Prescriptor.Prescriptor = pointer.FromString(common.HybridPrescriptor)
						mutator(datum)
						expectedDatum := CloneFood(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *food.Food) {},
				),
				Entry("does not modify the datum; amount missing",
					func(datum *food.Food) { datum.Amount = nil },
				),
				Entry("does not modify the datum; brand missing",
					func(datum *food.Food) { datum.Brand = nil },
				),
				Entry("does not modify the datum; code missing",
					func(datum *food.Food) { datum.Code = nil },
				),
				Entry("does not modify the datum; ingredients missing",
					func(datum *food.Food) { datum.Ingredients = nil },
				),
				Entry("does not modify the datum; meal missing",
					func(datum *food.Food) { datum.Meal = nil },
				),
				Entry("does not modify the datum; meal other missing",
					func(datum *food.Food) { datum.MealOther = nil },
				),
				Entry("does not modify the datum; name missing",
					func(datum *food.Food) { datum.Name = nil },
				),
				Entry("does not modify the datum; nutrition missing",
					func(datum *food.Food) { datum.Nutrition = nil },
				),
				Entry("does not modify the datum; prescribedNutrition missing",
					func(datum *food.Food) { datum.PrescribedNutrition = nil },
				),
			)

			DescribeTable("normalizes the datum for rescueCarbs and any other prescriptor; PrescribedNutrition is changed to nil",
				func(mutator func(datum *food.Food)) {
					for _, origin := range structure.Origins() {
						datum := NewFood(3)
						datum.Meal = pointer.FromString(food.MealRescueCarbs)
						datum.Prescriptor.Prescriptor = pointer.FromString(test.RandomStringFromArray([]string{common.ManualPrescriptor, common.AutoPrescriptor}))
						mutator(datum)
						expectedDatum := CloneFood(datum)
						expectedDatum.PrescribedNutrition = nil
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *food.Food) {},
				),
				Entry("does not modify the datum; amount missing",
					func(datum *food.Food) { datum.Amount = nil },
				),
				Entry("does not modify the datum; brand missing",
					func(datum *food.Food) { datum.Brand = nil },
				),
				Entry("does not modify the datum; code missing",
					func(datum *food.Food) { datum.Code = nil },
				),
				Entry("does not modify the datum; ingredients missing",
					func(datum *food.Food) { datum.Ingredients = nil },
				),
				Entry("does not modify the datum; meal missing",
					func(datum *food.Food) { datum.Meal = nil },
				),
				Entry("does not modify the datum; meal other missing",
					func(datum *food.Food) { datum.MealOther = nil },
				),
				Entry("does not modify the datum; name missing",
					func(datum *food.Food) { datum.Name = nil },
				),
				Entry("does not modify the datum; nutrition missing",
					func(datum *food.Food) { datum.Nutrition = nil },
				),
				Entry("modify the datum; prescribedNutrition is reset to nil",
					func(datum *food.Food) {
						datum.PrescribedNutrition = NewNutrition()
					},
				),
				Entry("does not crash when prescriptor is not set",
					func(datum *food.Food) {
						datum.Prescriptor.Prescriptor = nil
					},
				),
			)

		})
	})
})
