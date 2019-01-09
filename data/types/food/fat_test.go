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

func NewFat() *food.Fat {
	datum := food.NewFat()
	datum.Total = pointer.FromFloat64(test.RandomFloat64FromRange(food.FatTotalGramsMinimum, food.FatTotalGramsMaximum))
	datum.Units = pointer.FromString(test.RandomStringFromArray(food.FatUnits()))
	return datum
}

func CloneFat(datum *food.Fat) *food.Fat {
	if datum == nil {
		return nil
	}
	clone := food.NewFat()
	clone.Total = pointer.CloneFloat64(datum.Total)
	clone.Units = pointer.CloneString(datum.Units)
	return clone
}

var _ = Describe("Fat", func() {
	It("FatTotalGramsMaximum is expected", func() {
		Expect(food.FatTotalGramsMaximum).To(Equal(1000.0))
	})

	It("FatTotalGramsMinimum is expected", func() {
		Expect(food.FatTotalGramsMinimum).To(Equal(0.0))
	})

	It("FatUnitsGrams is expected", func() {
		Expect(food.FatUnitsGrams).To(Equal("grams"))
	})

	It("FatUnits returns expected", func() {
		Expect(food.FatUnits()).To(Equal([]string{"grams"}))
	})

	Context("ParseFat", func() {
		// TODO
	})

	Context("NewFat", func() {
		It("is successful", func() {
			Expect(food.NewFat()).To(Equal(&food.Fat{}))
		})
	})

	Context("Fat", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *food.Fat), expectedErrors ...error) {
					datum := NewFat()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *food.Fat) {},
				),
				Entry("total missing",
					func(datum *food.Fat) { datum.Total = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/total"),
				),
				Entry("total out of range (lower)",
					func(datum *food.Fat) { datum.Total = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/total"),
				),
				Entry("total in range (lower)",
					func(datum *food.Fat) { datum.Total = pointer.FromFloat64(0.0) },
				),
				Entry("total in range (upper)",
					func(datum *food.Fat) { datum.Total = pointer.FromFloat64(1000.0) },
				),
				Entry("total out of range (upper)",
					func(datum *food.Fat) { datum.Total = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/total"),
				),
				Entry("units missing",
					func(datum *food.Fat) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *food.Fat) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"grams"}), "/units"),
				),
				Entry("units grams",
					func(datum *food.Fat) { datum.Units = pointer.FromString("grams") },
				),
				Entry("multiple errors",
					func(datum *food.Fat) {
						datum.Total = nil
						datum.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/total"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *food.Fat)) {
					for _, origin := range structure.Origins() {
						datum := NewFat()
						mutator(datum)
						expectedDatum := CloneFat(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *food.Fat) {},
				),
				Entry("does not modify the datum; total missing",
					func(datum *food.Fat) { datum.Total = nil },
				),
				Entry("does not modify the datum; units missing",
					func(datum *food.Fat) { datum.Units = nil },
				),
			)
		})
	})
})
