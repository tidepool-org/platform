package food_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/food"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "food",
	}
}

func NewFood() *food.Food {
	datum := food.New()
	datum.Base = *testDataTypes.NewBase()
	datum.Type = "food"
	datum.Nutrition = NewNutrition()
	return datum
}

func CloneFood(datum *food.Food) *food.Food {
	if datum == nil {
		return nil
	}
	clone := food.New()
	clone.Base = *testDataTypes.CloneBase(&datum.Base)
	clone.Nutrition = CloneNutrition(datum.Nutrition)
	return clone
}

var _ = Describe("Food", func() {
	Context("Type", func() {
		It("returns the expected type", func() {
			Expect(food.Type()).To(Equal("food"))
		})
	})

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(food.NewDatum()).To(Equal(&food.Food{}))
		})
	})

	Context("New", func() {
		It("returns the expected datum", func() {
			Expect(food.New()).To(Equal(&food.Food{}))
		})
	})

	Context("Init", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := food.Init()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("food"))
			Expect(datum.Nutrition).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var datum *food.Food

		BeforeEach(func() {
			datum = NewFood()
		})

		Context("Init", func() {
			It("initializes the datum", func() {
				datum.Init()
				Expect(datum.Type).To(Equal("food"))
				Expect(datum.Nutrition).To(BeNil())
			})
		})
	})

	Context("Food", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *food.Food), expectedErrors ...error) {
					datum := NewFood()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *food.Food) {},
				),
				Entry("type missing",
					func(datum *food.Food) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				),
				Entry("type invalid",
					func(datum *food.Food) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "food"), "/type", &types.Meta{Type: "invalidType"}),
				),
				Entry("type food",
					func(datum *food.Food) { datum.Type = "food" },
				),
				Entry("nutrition missing",
					func(datum *food.Food) { datum.Nutrition = nil },
				),
				Entry("nutrition invalid",
					func(datum *food.Food) { datum.Nutrition.Carbohydrates.Net = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/nutrition/carbohydrates/net", NewMeta()),
				),
				Entry("nutrition valid",
					func(datum *food.Food) { datum.Nutrition = NewNutrition() },
				),
				Entry("multiple errors",
					func(datum *food.Food) {
						datum.Type = "invalidType"
						datum.Nutrition.Carbohydrates.Net = nil
						datum.Nutrition.Carbohydrates.Units = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "food"), "/type", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/nutrition/carbohydrates/net", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/nutrition/carbohydrates/units", &types.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *food.Food)) {
					for _, origin := range structure.Origins() {
						datum := NewFood()
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
				Entry("does not modify the datum; nutrition missing",
					func(datum *food.Food) { datum.Nutrition = nil },
				),
			)
		})
	})
})
