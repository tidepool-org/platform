package food_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/food"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewCarbohydrates() *food.Carbohydrates {
	datum := food.NewCarbohydrates()
	datum.Net = pointer.Int(test.RandomIntFromRange(food.ValueGramsMinimum, food.ValueGramsMaximum))
	datum.Units = pointer.String(test.RandomStringFromArray(food.Units()))
	return datum
}

func CloneCarbohydrates(datum *food.Carbohydrates) *food.Carbohydrates {
	if datum == nil {
		return nil
	}
	clone := food.NewCarbohydrates()
	clone.Net = test.CloneInt(datum.Net)
	clone.Units = test.CloneString(datum.Units)
	return clone
}

var _ = Describe("Carbohydrates", func() {
	It("UnitsGrams is expected", func() {
		Expect(food.UnitsGrams).To(Equal("grams"))
	})

	It("ValueGramsMaximum is expected", func() {
		Expect(food.ValueGramsMaximum).To(Equal(1000))
	})

	It("ValueGramsMinimum is expected", func() {
		Expect(food.ValueGramsMinimum).To(Equal(0))
	})

	It("Units returns expected", func() {
		Expect(food.Units()).To(Equal([]string{"grams"}))
	})

	Context("ParseCarbohydrates", func() {
		// TODO
	})

	Context("NewCarbohydrates", func() {
		It("is successful", func() {
			Expect(food.NewCarbohydrates()).To(Equal(&food.Carbohydrates{}))
		})
	})

	Context("Carbohydrates", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *food.Carbohydrates), expectedErrors ...error) {
					datum := NewCarbohydrates()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *food.Carbohydrates) {},
				),
				Entry("net missing",
					func(datum *food.Carbohydrates) { datum.Net = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/net"),
				),
				Entry("net out of range (lower)",
					func(datum *food.Carbohydrates) { datum.Net = pointer.Int(-1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 1000), "/net"),
				),
				Entry("net in range (lower)",
					func(datum *food.Carbohydrates) { datum.Net = pointer.Int(0) },
				),
				Entry("net in range (upper)",
					func(datum *food.Carbohydrates) { datum.Net = pointer.Int(1000) },
				),
				Entry("net out of range (upper)",
					func(datum *food.Carbohydrates) { datum.Net = pointer.Int(1001) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1001, 0, 1000), "/net"),
				),
				Entry("units missing",
					func(datum *food.Carbohydrates) { datum.Units = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *food.Carbohydrates) { datum.Units = pointer.String("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"grams"}), "/units"),
				),
				Entry("units grams",
					func(datum *food.Carbohydrates) { datum.Units = pointer.String("grams") },
				),
				Entry("multiple errors",
					func(datum *food.Carbohydrates) {
						datum.Net = nil
						datum.Units = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/net"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *food.Carbohydrates)) {
					for _, origin := range structure.Origins() {
						datum := NewCarbohydrates()
						mutator(datum)
						expectedDatum := CloneCarbohydrates(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *food.Carbohydrates) {},
				),
				Entry("does not modify the datum; net missing",
					func(datum *food.Carbohydrates) { datum.Net = nil },
				),
				Entry("does not modify the datum; units missing",
					func(datum *food.Carbohydrates) { datum.Units = nil },
				),
			)
		})
	})
})
