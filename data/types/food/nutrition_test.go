package food_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/food"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewNutrition() *food.Nutrition {
	datum := food.NewNutrition()
	datum.EstimatedAbsorptionDuration = pointer.FromInt(test.RandomIntFromRange(food.EstimatedAbsorptionDurationSecondsMinimum, food.EstimatedAbsorptionDurationSecondsMaximum))
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
	clone.EstimatedAbsorptionDuration = pointer.CloneInt(datum.EstimatedAbsorptionDuration)
	clone.Carbohydrate = CloneCarbohydrate(datum.Carbohydrate)
	clone.Energy = CloneEnergy(datum.Energy)
	clone.Fat = CloneFat(datum.Fat)
	clone.Protein = CloneProtein(datum.Protein)
	return clone
}

var _ = Describe("Nutrition", func() {
	It("EstimatedAbsorptionDuration is expected", func() {
		Expect(food.EstimatedAbsorptionDurationSecondsMaximum).To(Equal(86400))
	})

	It("AbsorptionDurationSecondsMinimum is expected", func() {
		Expect(food.EstimatedAbsorptionDurationSecondsMinimum).To(Equal(0))
	})

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
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *food.Nutrition) {},
				),
				Entry("absorption duration missing",
					func(datum *food.Nutrition) { datum.EstimatedAbsorptionDuration = nil },
				),
				Entry("absorption duration out of range (lower)",
					func(datum *food.Nutrition) { datum.EstimatedAbsorptionDuration = pointer.FromInt(-1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 86400), "/absorptionDuration"),
				),
				Entry("absorption duration in range (lower)",
					func(datum *food.Nutrition) { datum.EstimatedAbsorptionDuration = pointer.FromInt(0) },
				),
				Entry("absorption duration in range (upper)",
					func(datum *food.Nutrition) { datum.EstimatedAbsorptionDuration = pointer.FromInt(86400) },
				),
				Entry("absorption duration out of range (upper)",
					func(datum *food.Nutrition) { datum.EstimatedAbsorptionDuration = pointer.FromInt(86401) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(86401, 0, 86400), "/absorptionDuration"),
				),
				Entry("carbohydrate missing",
					func(datum *food.Nutrition) { datum.Carbohydrate = nil },
				),
				Entry("carbohydrate invalid",
					func(datum *food.Nutrition) { datum.Carbohydrate.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/carbohydrate/units"),
				),
				Entry("carbohydrate valid",
					func(datum *food.Nutrition) { datum.Carbohydrate = NewCarbohydrate() },
				),
				Entry("energy missing",
					func(datum *food.Nutrition) { datum.Energy = nil },
				),
				Entry("energy invalid",
					func(datum *food.Nutrition) { datum.Energy.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/energy/units"),
				),
				Entry("energy valid",
					func(datum *food.Nutrition) { datum.Energy = NewEnergy() },
				),
				Entry("fat missing",
					func(datum *food.Nutrition) { datum.Fat = nil },
				),
				Entry("fat invalid",
					func(datum *food.Nutrition) { datum.Fat.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/fat/units"),
				),
				Entry("fat valid",
					func(datum *food.Nutrition) { datum.Fat = NewFat() },
				),
				Entry("protein missing",
					func(datum *food.Nutrition) { datum.Protein = nil },
				),
				Entry("protein invalid",
					func(datum *food.Nutrition) { datum.Protein.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/protein/units"),
				),
				Entry("protein valid",
					func(datum *food.Nutrition) { datum.Protein = NewProtein() },
				),
				Entry("multiple errors",
					func(datum *food.Nutrition) {
						datum.EstimatedAbsorptionDuration = pointer.FromInt(-1)
						datum.Carbohydrate.Units = nil
						datum.Energy.Units = nil
						datum.Fat.Units = nil
						datum.Protein.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 86400), "/absorptionDuration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/carbohydrate/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/energy/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/fat/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/protein/units"),
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
				Entry("does not modify the datum; absorption duration missing",
					func(datum *food.Nutrition) { datum.EstimatedAbsorptionDuration = nil },
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
