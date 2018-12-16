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

func NewCarbohydrate() *food.Carbohydrate {
	datum := food.NewCarbohydrate()
	datum.DietaryFiber = pointer.FromFloat64(test.RandomFloat64FromRange(food.CarbohydrateDietaryFiberGramsMinimum, food.CarbohydrateDietaryFiberGramsMaximum))
	datum.Net = pointer.FromFloat64(test.RandomFloat64FromRange(food.CarbohydrateNetGramsMinimum, food.CarbohydrateNetGramsMaximum))
	datum.Sugars = pointer.FromFloat64(test.RandomFloat64FromRange(food.CarbohydrateSugarsGramsMinimum, food.CarbohydrateSugarsGramsMaximum))
	datum.Total = pointer.FromFloat64(test.RandomFloat64FromRange(food.CarbohydrateTotalGramsMinimum, food.CarbohydrateTotalGramsMaximum))
	datum.Units = pointer.FromString(test.RandomStringFromArray(food.CarbohydrateUnits()))
	return datum
}

func CloneCarbohydrate(datum *food.Carbohydrate) *food.Carbohydrate {
	if datum == nil {
		return nil
	}
	clone := food.NewCarbohydrate()
	clone.DietaryFiber = test.CloneFloat64(datum.DietaryFiber)
	clone.Net = test.CloneFloat64(datum.Net)
	clone.Sugars = test.CloneFloat64(datum.Sugars)
	clone.Total = test.CloneFloat64(datum.Total)
	clone.Units = test.CloneString(datum.Units)
	return clone
}

var _ = Describe("Carbohydrate", func() {
	It("CarbohydrateDietaryFiberGramsMaximum is expected", func() {
		Expect(food.CarbohydrateDietaryFiberGramsMaximum).To(Equal(1000.0))
	})

	It("CarbohydrateDietaryFiberGramsMinimum is expected", func() {
		Expect(food.CarbohydrateDietaryFiberGramsMinimum).To(Equal(0.0))
	})

	It("CarbohydrateNetGramsMaximum is expected", func() {
		Expect(food.CarbohydrateNetGramsMaximum).To(Equal(1000.0))
	})

	It("CarbohydrateNetGramsMinimum is expected", func() {
		Expect(food.CarbohydrateNetGramsMinimum).To(Equal(0.0))
	})

	It("CarbohydrateSugarsGramsMaximum is expected", func() {
		Expect(food.CarbohydrateSugarsGramsMaximum).To(Equal(1000.0))
	})

	It("CarbohydrateSugarsGramsMinimum is expected", func() {
		Expect(food.CarbohydrateSugarsGramsMinimum).To(Equal(0.0))
	})

	It("CarbohydrateTotalGramsMaximum is expected", func() {
		Expect(food.CarbohydrateTotalGramsMaximum).To(Equal(1000.0))
	})

	It("CarbohydrateTotalGramsMinimum is expected", func() {
		Expect(food.CarbohydrateTotalGramsMinimum).To(Equal(0.0))
	})

	It("CarbohydrateUnitsGrams is expected", func() {
		Expect(food.CarbohydrateUnitsGrams).To(Equal("grams"))
	})

	It("CarbohydrateUnits returns expected", func() {
		Expect(food.CarbohydrateUnits()).To(Equal([]string{"grams"}))
	})

	Context("ParseCarbohydrate", func() {
		// TODO
	})

	Context("NewCarbohydrate", func() {
		It("is successful", func() {
			Expect(food.NewCarbohydrate()).To(Equal(&food.Carbohydrate{}))
		})
	})

	Context("Carbohydrate", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *food.Carbohydrate), expectedErrors ...error) {
					datum := NewCarbohydrate()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *food.Carbohydrate) {},
				),
				Entry("dietary fiber missing",
					func(datum *food.Carbohydrate) { datum.DietaryFiber = nil },
				),
				Entry("dietary fiber out of range (lower)",
					func(datum *food.Carbohydrate) { datum.DietaryFiber = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/dietaryFiber"),
				),
				Entry("dietary fiber in range (lower)",
					func(datum *food.Carbohydrate) { datum.DietaryFiber = pointer.FromFloat64(0.0) },
				),
				Entry("dietary fiber in range (upper)",
					func(datum *food.Carbohydrate) { datum.DietaryFiber = pointer.FromFloat64(1000.0) },
				),
				Entry("dietary fiber out of range (upper)",
					func(datum *food.Carbohydrate) { datum.DietaryFiber = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/dietaryFiber"),
				),
				Entry("net missing",
					func(datum *food.Carbohydrate) { datum.Net = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/net"),
				),
				Entry("net out of range (lower)",
					func(datum *food.Carbohydrate) { datum.Net = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/net"),
				),
				Entry("net in range (lower)",
					func(datum *food.Carbohydrate) { datum.Net = pointer.FromFloat64(0.0) },
				),
				Entry("net in range (upper)",
					func(datum *food.Carbohydrate) { datum.Net = pointer.FromFloat64(1000.0) },
				),
				Entry("net out of range (upper)",
					func(datum *food.Carbohydrate) { datum.Net = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/net"),
				),
				Entry("sugars missing",
					func(datum *food.Carbohydrate) { datum.Sugars = nil },
				),
				Entry("sugars out of range (lower)",
					func(datum *food.Carbohydrate) { datum.Sugars = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/sugars"),
				),
				Entry("sugars in range (lower)",
					func(datum *food.Carbohydrate) { datum.Sugars = pointer.FromFloat64(0.0) },
				),
				Entry("sugars in range (upper)",
					func(datum *food.Carbohydrate) { datum.Sugars = pointer.FromFloat64(1000.0) },
				),
				Entry("sugars out of range (upper)",
					func(datum *food.Carbohydrate) { datum.Sugars = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/sugars"),
				),
				Entry("total missing",
					func(datum *food.Carbohydrate) { datum.Total = nil },
				),
				Entry("total out of range (lower)",
					func(datum *food.Carbohydrate) { datum.Total = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/total"),
				),
				Entry("total in range (lower)",
					func(datum *food.Carbohydrate) { datum.Total = pointer.FromFloat64(0.0) },
				),
				Entry("total in range (upper)",
					func(datum *food.Carbohydrate) { datum.Total = pointer.FromFloat64(1000.0) },
				),
				Entry("total out of range (upper)",
					func(datum *food.Carbohydrate) { datum.Total = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/total"),
				),
				Entry("units missing",
					func(datum *food.Carbohydrate) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *food.Carbohydrate) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"grams"}), "/units"),
				),
				Entry("units grams",
					func(datum *food.Carbohydrate) { datum.Units = pointer.FromString("grams") },
				),
				Entry("multiple errors",
					func(datum *food.Carbohydrate) {
						datum.DietaryFiber = pointer.FromFloat64(-0.1)
						datum.Net = pointer.FromFloat64(-0.1)
						datum.Sugars = pointer.FromFloat64(-0.1)
						datum.Total = pointer.FromFloat64(-0.1)
						datum.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/dietaryFiber"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/net"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/sugars"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/total"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *food.Carbohydrate)) {
					for _, origin := range structure.Origins() {
						datum := NewCarbohydrate()
						mutator(datum)
						expectedDatum := CloneCarbohydrate(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *food.Carbohydrate) {},
				),
				Entry("does not modify the datum; dietary fiber missing",
					func(datum *food.Carbohydrate) { datum.DietaryFiber = nil },
				),
				Entry("does not modify the datum; net missing",
					func(datum *food.Carbohydrate) { datum.Net = nil },
				),
				Entry("does not modify the datum; sugars missing",
					func(datum *food.Carbohydrate) { datum.Sugars = nil },
				),
				Entry("does not modify the datum; total missing",
					func(datum *food.Carbohydrate) { datum.Total = nil },
				),
				Entry("does not modify the datum; units missing",
					func(datum *food.Carbohydrate) { datum.Units = nil },
				),
			)
		})
	})
})
