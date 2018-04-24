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
	datum.Carbohydrates = NewCarbohydrates()
	return datum
}

func CloneNutrition(datum *food.Nutrition) *food.Nutrition {
	if datum == nil {
		return nil
	}
	clone := food.NewNutrition()
	clone.Carbohydrates = CloneCarbohydrates(datum.Carbohydrates)
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
				Entry("carbohydrates missing",
					func(datum *food.Nutrition) { datum.Carbohydrates = nil },
				),
				Entry("carbohydrates invalid",
					func(datum *food.Nutrition) { datum.Carbohydrates.Net = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/carbohydrates/net"),
				),
				Entry("carbohydrates valid",
					func(datum *food.Nutrition) { datum.Carbohydrates = NewCarbohydrates() },
				),
				Entry("multiple errors",
					func(datum *food.Nutrition) {
						datum.Carbohydrates.Net = nil
						datum.Carbohydrates.Units = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/carbohydrates/net"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/carbohydrates/units"),
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
				Entry("does not modify the datum; carbohydrates missing",
					func(datum *food.Nutrition) { datum.Carbohydrates = nil },
				),
			)
		})
	})
})
