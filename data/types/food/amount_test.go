package food_test

import (
	"math"

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

func NewAmount() *food.Amount {
	datum := food.NewAmount()
	datum.Units = pointer.FromString(test.NewText(1, 100))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(0.0, math.MaxFloat64))
	return datum
}

func CloneAmount(datum *food.Amount) *food.Amount {
	if datum == nil {
		return nil
	}
	clone := food.NewAmount()
	clone.Units = test.CloneString(datum.Units)
	clone.Value = test.CloneFloat64(datum.Value)
	return clone
}

var _ = Describe("Amount", func() {
	It("AmountUnitsLengthMaximum is expected", func() {
		Expect(food.AmountUnitsLengthMaximum).To(Equal(100))
	})

	It("AmountValueMinimum is expected", func() {
		Expect(food.AmountValueMinimum).To(Equal(0.0))
	})

	Context("ParseAmount", func() {
		// TODO
	})

	Context("NewAmount", func() {
		It("is successful", func() {
			Expect(food.NewAmount()).To(Equal(&food.Amount{}))
		})
	})

	Context("Amount", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *food.Amount), expectedErrors ...error) {
					datum := NewAmount()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *food.Amount) {},
				),
				Entry("units missing",
					func(datum *food.Amount) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units empty",
					func(datum *food.Amount) { datum.Units = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/units"),
				),
				Entry("units invalid",
					func(datum *food.Amount) { datum.Units = pointer.FromString(test.NewText(101, 101)) },
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/units"),
				),
				Entry("units valid",
					func(datum *food.Amount) { datum.Units = pointer.FromString(test.NewText(1, 100)) },
				),
				Entry("value missing",
					func(datum *food.Amount) { datum.Value = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("value out of range (lower)",
					func(datum *food.Amount) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-0.1, 0.0), "/value"),
				),
				Entry("value in range (lower)",
					func(datum *food.Amount) { datum.Value = pointer.FromFloat64(0.0) },
				),
				Entry("value in range (upper)",
					func(datum *food.Amount) { datum.Value = pointer.FromFloat64(math.MaxFloat64) },
				),
				Entry("multiple errors",
					func(datum *food.Amount) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *food.Amount)) {
					for _, origin := range structure.Origins() {
						datum := NewAmount()
						mutator(datum)
						expectedDatum := CloneAmount(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *food.Amount) {},
				),
				Entry("does not modify the datum; units missing",
					func(datum *food.Amount) { datum.Units = nil },
				),
				Entry("does not modify the datum; value missing",
					func(datum *food.Amount) { datum.Value = nil },
				),
			)
		})
	})
})
