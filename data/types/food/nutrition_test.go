package food_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/food"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewNutrition() *food.Nutrition {
	datum := food.NewNutrition()
	datum.Carbohydrate = NewCarbohydrate()
	datum.Energy = NewEnergy()
	datum.Fat = NewFat()
	datum.Protein = NewProtein()
	return datum
}

func CloneNutrition(datum *food.Nutrition) *food.Nutrition {
	if datum == nil {
		return nil
	}
	clone := food.NewNutrition()
	clone.Carbohydrate = CloneCarbohydrate(datum.Carbohydrate)
	clone.Energy = CloneEnergy(datum.Energy)
	clone.Fat = CloneFat(datum.Fat)
	clone.Protein = CloneProtein(datum.Protein)
	return clone
}

var _ = Describe("Nutrition", func() {
	Context("ParseNutrition", func() {
		// TODO
	})

	Context("NewNutrition", func() {
		It("is successful", func() {
			Expect(food.NewNutrition()).To(Equal(&food.Nutrition{}))
		})
	})

	Context("Nutrition", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *food.Nutrition), expectedErrors ...error) {
					datum := NewNutrition()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *food.Nutrition) {},
				),
				Entry("carbohydrate missing",
					func(datum *food.Nutrition) { datum.Carbohydrate = nil },
				),
				Entry("carbohydrate invalid",
					func(datum *food.Nutrition) { datum.Carbohydrate.Units = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/carbohydrate/units"),
				),
				Entry("carbohydrate valid",
					func(datum *food.Nutrition) { datum.Carbohydrate = NewCarbohydrate() },
				),
				Entry("energy missing",
					func(datum *food.Nutrition) { datum.Energy = nil },
				),
				Entry("energy invalid",
					func(datum *food.Nutrition) { datum.Energy.Units = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/energy/units"),
				),
				Entry("energy valid",
					func(datum *food.Nutrition) { datum.Energy = NewEnergy() },
				),
				Entry("fat missing",
					func(datum *food.Nutrition) { datum.Fat = nil },
				),
				Entry("fat invalid",
					func(datum *food.Nutrition) { datum.Fat.Units = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/fat/units"),
				),
				Entry("fat valid",
					func(datum *food.Nutrition) { datum.Fat = NewFat() },
				),
				Entry("protein missing",
					func(datum *food.Nutrition) { datum.Protein = nil },
				),
				Entry("protein invalid",
					func(datum *food.Nutrition) { datum.Protein.Units = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/protein/units"),
				),
				Entry("protein valid",
					func(datum *food.Nutrition) { datum.Protein = NewProtein() },
				),
				Entry("multiple errors",
					func(datum *food.Nutrition) {
						datum.Carbohydrate.Units = nil
						datum.Energy.Units = nil
						datum.Fat.Units = nil
						datum.Protein.Units = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/carbohydrate/units"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/energy/units"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/fat/units"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/protein/units"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *food.Nutrition)) {
					for _, origin := range structure.Origins() {
						datum := NewNutrition()
						mutator(datum)
						expectedDatum := CloneNutrition(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *food.Nutrition) {},
				),
				Entry("does not modify the datum; carbohydrate missing",
					func(datum *food.Nutrition) { datum.Carbohydrate = nil },
				),
				Entry("does not modify the datum; energy missing",
					func(datum *food.Nutrition) { datum.Energy = nil },
				),
				Entry("does not modify the datum; fat missing",
					func(datum *food.Nutrition) { datum.Fat = nil },
				),
				Entry("does not modify the datum; protein missing",
					func(datum *food.Nutrition) { datum.Protein = nil },
				),
			)
		})
	})
})
