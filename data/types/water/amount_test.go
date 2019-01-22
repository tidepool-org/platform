package water_test

import (
	"math"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/data/types/water"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewAmount() *water.Amount {
	datum := water.NewAmount()
	datum.Units = pointer.FromString(test.RandomStringFromArray(water.AmountUnits()))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(water.AmountValueRangeForUnits(datum.Units)))
	return datum
}

func CloneAmount(datum *water.Amount) *water.Amount {
	if datum == nil {
		return nil
	}
	clone := water.NewAmount()
	clone.Units = pointer.CloneString(datum.Units)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}

var _ = Describe("Amount", func() {
	It("AmountLitersPerGallon is expected", func() {
		Expect(water.AmountLitersPerGallon).To(Equal(3.7854118))
	})

	It("AmountOuncesPerGallon is expected", func() {
		Expect(water.AmountOuncesPerGallon).To(Equal(128.0))
	})

	It("AmountUnitsGallons is expected", func() {
		Expect(water.AmountUnitsGallons).To(Equal("gallons"))
	})

	It("AmountUnitsLiters is expected", func() {
		Expect(water.AmountUnitsLiters).To(Equal("liters"))
	})

	It("AmountUnitsMilliliters is expected", func() {
		Expect(water.AmountUnitsMilliliters).To(Equal("milliliters"))
	})

	It("AmountUnitsOunces is expected", func() {
		Expect(water.AmountUnitsOunces).To(Equal("ounces"))
	})

	It("AmountValueGallonsMaximum is expected", func() {
		Expect(water.AmountValueGallonsMaximum).To(Equal(10.0))
	})

	It("AmountValueGallonsMinimum is expected", func() {
		Expect(water.AmountValueGallonsMinimum).To(Equal(0.0))
	})

	It("AmountValueLitersMaximum is expected", func() {
		Expect(water.AmountValueLitersMaximum).To(Equal(37.854118))
	})

	It("AmountValueLitersMinimum is expected", func() {
		Expect(water.AmountValueLitersMinimum).To(Equal(0.0))
	})

	It("AmountValueMillilitersMaximum is expected", func() {
		Expect(water.AmountValueMillilitersMaximum).To(Equal(37854.118))
	})

	It("AmountValueMillilitersMinimum is expected", func() {
		Expect(water.AmountValueMillilitersMinimum).To(Equal(0.0))
	})

	It("AmountValueOuncesMaximum is expected", func() {
		Expect(water.AmountValueOuncesMaximum).To(Equal(1280.0))
	})

	It("AmountValueOuncesMinimum is expected", func() {
		Expect(water.AmountValueOuncesMinimum).To(Equal(0.0))
	})

	It("AmountUnits returns expected", func() {
		Expect(water.AmountUnits()).To(Equal([]string{"gallons", "liters", "milliliters", "ounces"}))
	})

	Context("ParseAmount", func() {
		// TODO
	})

	Context("NewAmount", func() {
		It("is successful", func() {
			Expect(water.NewAmount()).To(Equal(&water.Amount{}))
		})
	})

	Context("Amount", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *water.Amount), expectedErrors ...error) {
					datum := NewAmount()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *water.Amount) {},
				),
				Entry("units missing",
					func(datum *water.Amount) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *water.Amount) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"gallons", "liters", "milliliters", "ounces"}), "/units"),
				),
				Entry("units gallons",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("gallons")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units liters",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("liters")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units milliliters",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("milliliters")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units ounces",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("ounces")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units missing; value missing",
					func(datum *water.Amount) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value out of range (lower)",
					func(datum *water.Amount) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *water.Amount) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *water.Amount) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(37854.118)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value out of range (upper)",
					func(datum *water.Amount) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(37854.119)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"gallons", "liters", "milliliters", "ounces"}), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value out of range (lower)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"gallons", "liters", "milliliters", "ounces"}), "/units"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"gallons", "liters", "milliliters", "ounces"}), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(37854.118)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"gallons", "liters", "milliliters", "ounces"}), "/units"),
				),
				Entry("units invalid; value out of range (upper)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(37854.119)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"gallons", "liters", "milliliters", "ounces"}), "/units"),
				),
				Entry("units gallons; value missing",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("gallons")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units gallons; value out of range (lower)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("gallons")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10.0), "/value"),
				),
				Entry("units gallons; value in range (lower)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("gallons")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units gallons; value in range (upper)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("gallons")
						datum.Value = pointer.FromFloat64(10.0)
					},
				),
				Entry("units gallons; value out of range (upper)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("gallons")
						datum.Value = pointer.FromFloat64(10.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10.1, 0.0, 10.0), "/value"),
				),
				Entry("units liters; value missing",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("liters")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units liters; value out of range (lower)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("liters")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 37.854118), "/value"),
				),
				Entry("units liters; value in range (lower)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("liters")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units liters; value in range (upper)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("liters")
						datum.Value = pointer.FromFloat64(37.854118)
					},
				),
				Entry("units liters; value out of range (upper)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("liters")
						datum.Value = pointer.FromFloat64(37.854119)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(37.854119, 0.0, 37.854118), "/value"),
				),
				Entry("units milliliters; value missing",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("milliliters")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units milliliters; value out of range (lower)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("milliliters")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 37854.118), "/value"),
				),
				Entry("units milliliters; value in range (lower)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("milliliters")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units milliliters; value in range (upper)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("milliliters")
						datum.Value = pointer.FromFloat64(37854.118)
					},
				),
				Entry("units milliliters; value out of range (upper)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("milliliters")
						datum.Value = pointer.FromFloat64(37854.119)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(37854.119, 0.0, 37854.118), "/value"),
				),
				Entry("units ounces; value missing",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("ounces")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units ounces; value out of range (lower)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("ounces")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1280.0), "/value"),
				),
				Entry("units ounces; value in range (lower)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("ounces")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units ounces; value in range (upper)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("ounces")
						datum.Value = pointer.FromFloat64(1280.0)
					},
				),
				Entry("units ounces; value out of range (upper)",
					func(datum *water.Amount) {
						datum.Units = pointer.FromString("ounces")
						datum.Value = pointer.FromFloat64(1280.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1280.1, 0.0, 1280.0), "/value"),
				),
				Entry("multiple errors",
					func(datum *water.Amount) {
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
				func(mutator func(datum *water.Amount)) {
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
					func(datum *water.Amount) {},
				),
				Entry("does not modify the datum; units missing",
					func(datum *water.Amount) { datum.Units = nil },
				),
				Entry("does not modify the datum; value missing",
					func(datum *water.Amount) { datum.Value = nil },
				),
			)
		})
	})

	Context("AmountValueRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := water.AmountValueRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := water.AmountValueRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units gallons", func() {
			minimum, maximum := water.AmountValueRangeForUnits(pointer.FromString("gallons"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(10.0))
		})

		It("returns expected range for units liters", func() {
			minimum, maximum := water.AmountValueRangeForUnits(pointer.FromString("liters"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(37.854118))
		})

		It("returns expected range for units milliliters", func() {
			minimum, maximum := water.AmountValueRangeForUnits(pointer.FromString("milliliters"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(37854.118))
		})

		It("returns expected range for units ounces", func() {
			minimum, maximum := water.AmountValueRangeForUnits(pointer.FromString("ounces"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(1280.0))
		})
	})
})
