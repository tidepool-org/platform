package food_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/food"
	mealTest "github.com/tidepool-org/platform/data/types/food/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Meal", func() {
	Context("NewMeal", func() {
		It("is successful", func() {
			Expect(food.NewMeal()).To(Equal(&food.Meal{}))
		})
	})

	Context("Meal", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *food.Meal), expectedErrors ...error) {
					datum := mealTest.NewMeal()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *food.Meal) {},
				),
				Entry("meal invalid",
					func(datum *food.Meal) {
						datum.Meal = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"small", "medium", "large"}), "/meal"),
				),
				Entry("snack invalid",
					func(datum *food.Meal) {
						datum.Snack = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"yes", "no"}), "/snack"),
				),
				Entry("fat invalid",
					func(datum *food.Meal) {
						datum.Fat = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"yes", "no"}), "/fat"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *food.Meal)) {
					for _, origin := range structure.Origins() {
						datum := mealTest.NewMeal()
						mutator(datum)
						expectedDatum := mealTest.CloneMeal(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *food.Meal) {},
				),
			)
		})
	})
})
