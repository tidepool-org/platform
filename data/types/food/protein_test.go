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

func NewProtein() *food.Protein {
	datum := food.NewProtein()
	datum.Total = pointer.Int(test.RandomIntFromRange(food.ProteinTotalGramsMinimum, food.ProteinTotalGramsMaximum))
	datum.Units = pointer.String(test.RandomStringFromArray(food.ProteinUnits()))
	return datum
}

func CloneProtein(datum *food.Protein) *food.Protein {
	if datum == nil {
		return nil
	}
	clone := food.NewProtein()
	clone.Total = test.CloneInt(datum.Total)
	clone.Units = test.CloneString(datum.Units)
	return clone
}

var _ = Describe("Protein", func() {
	It("ProteinTotalGramsMaximum is expected", func() {
		Expect(food.ProteinTotalGramsMaximum).To(Equal(1000))
	})

	It("ProteinTotalGramsMinimum is expected", func() {
		Expect(food.ProteinTotalGramsMinimum).To(Equal(0))
	})

	It("ProteinUnitsGrams is expected", func() {
		Expect(food.ProteinUnitsGrams).To(Equal("grams"))
	})

	It("ProteinUnits returns expected", func() {
		Expect(food.ProteinUnits()).To(Equal([]string{"grams"}))
	})

	Context("ParseProtein", func() {
		// TODO
	})

	Context("NewProtein", func() {
		It("is successful", func() {
			Expect(food.NewProtein()).To(Equal(&food.Protein{}))
		})
	})

	Context("Protein", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *food.Protein), expectedErrors ...error) {
					datum := NewProtein()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *food.Protein) {},
				),
				Entry("total missing",
					func(datum *food.Protein) { datum.Total = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/total"),
				),
				Entry("total out of range (lower)",
					func(datum *food.Protein) { datum.Total = pointer.Int(-1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 1000), "/total"),
				),
				Entry("total in range (lower)",
					func(datum *food.Protein) { datum.Total = pointer.Int(0) },
				),
				Entry("total in range (upper)",
					func(datum *food.Protein) { datum.Total = pointer.Int(1000) },
				),
				Entry("total out of range (upper)",
					func(datum *food.Protein) { datum.Total = pointer.Int(1001) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1001, 0, 1000), "/total"),
				),
				Entry("units missing",
					func(datum *food.Protein) { datum.Units = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *food.Protein) { datum.Units = pointer.String("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"grams"}), "/units"),
				),
				Entry("units grams",
					func(datum *food.Protein) { datum.Units = pointer.String("grams") },
				),
				Entry("multiple errors",
					func(datum *food.Protein) {
						datum.Total = nil
						datum.Units = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/total"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *food.Protein)) {
					for _, origin := range structure.Origins() {
						datum := NewProtein()
						mutator(datum)
						expectedDatum := CloneProtein(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *food.Protein) {},
				),
				Entry("does not modify the datum; total missing",
					func(datum *food.Protein) { datum.Total = nil },
				),
				Entry("does not modify the datum; units missing",
					func(datum *food.Protein) { datum.Units = nil },
				),
			)
		})
	})
})
