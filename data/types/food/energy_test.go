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

func NewEnergy() *food.Energy {
	datum := food.NewEnergy()
	datum.Units = pointer.FromString(test.RandomStringFromArray(food.EnergyUnits()))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(food.EnergyValueRangeForUnits(datum.Units)))
	return datum
}

func CloneEnergy(datum *food.Energy) *food.Energy {
	if datum == nil {
		return nil
	}
	clone := food.NewEnergy()
	clone.Units = pointer.CloneString(datum.Units)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}

var _ = Describe("Energy", func() {
	It("EnergyKilojoulesPerKilocalorie is expected", func() {
		Expect(food.EnergyKilojoulesPerKilocalorie).To(Equal(4.1858))
	})

	It("EnergyUnitsCalories is expected", func() {
		Expect(food.EnergyUnitsCalories).To(Equal("calories"))
	})

	It("EnergyUnitsJoules is expected", func() {
		Expect(food.EnergyUnitsJoules).To(Equal("joules"))
	})

	It("EnergyUnitsKilocalories is expected", func() {
		Expect(food.EnergyUnitsKilocalories).To(Equal("kilocalories"))
	})

	It("EnergyUnitsKilojoules is expected", func() {
		Expect(food.EnergyUnitsKilojoules).To(Equal("kilojoules"))
	})

	It("EnergyValueCaloriesMaximum is expected", func() {
		Expect(food.EnergyValueCaloriesMaximum).To(Equal(10000000.0))
	})

	It("EnergyValueCaloriesMinimum is expected", func() {
		Expect(food.EnergyValueCaloriesMinimum).To(Equal(0.0))
	})

	It("EnergyValueJoulesMaximum is expected", func() {
		Expect(food.EnergyValueJoulesMaximum).To(Equal(41858000.0))
	})

	It("EnergyValueJoulesMinimum is expected", func() {
		Expect(food.EnergyValueJoulesMinimum).To(Equal(0.0))
	})

	It("EnergyValueKilocaloriesMaximum is expected", func() {
		Expect(food.EnergyValueKilocaloriesMaximum).To(Equal(10000.0))
	})

	It("EnergyValueKilocaloriesMinimum is expected", func() {
		Expect(food.EnergyValueKilocaloriesMinimum).To(Equal(0.0))
	})

	It("EnergyValueKilojoulesMaximum is expected", func() {
		Expect(food.EnergyValueKilojoulesMaximum).To(Equal(41858.0))
	})

	It("EnergyValueKilojoulesMinimum is expected", func() {
		Expect(food.EnergyValueKilojoulesMinimum).To(Equal(0.0))
	})

	It("EnergyUnits returns expected", func() {
		Expect(food.EnergyUnits()).To(Equal([]string{"calories", "joules", "kilocalories", "kilojoules"}))
	})

	Context("ParseEnergy", func() {
		// TODO
	})

	Context("NewEnergy", func() {
		It("is successful", func() {
			Expect(food.NewEnergy()).To(Equal(&food.Energy{}))
		})
	})

	Context("Energy", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *food.Energy), expectedErrors ...error) {
					datum := NewEnergy()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *food.Energy) {},
				),
				Entry("units missing",
					func(datum *food.Energy) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *food.Energy) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"calories", "joules", "kilocalories", "kilojoules"}), "/units"),
				),
				Entry("units calories",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("calories")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units joules",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("joules")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units kilocalories",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("kilocalories")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units kilojoules",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("kilojoules")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units missing; value missing",
					func(datum *food.Energy) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value out of range (lower)",
					func(datum *food.Energy) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *food.Energy) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *food.Energy) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(41858000.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value out of range (upper)",
					func(datum *food.Energy) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(41858000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"calories", "joules", "kilocalories", "kilojoules"}), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value out of range (lower)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"calories", "joules", "kilocalories", "kilojoules"}), "/units"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"calories", "joules", "kilocalories", "kilojoules"}), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(41858000.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"calories", "joules", "kilocalories", "kilojoules"}), "/units"),
				),
				Entry("units invalid; value out of range (upper)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(41858000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"calories", "joules", "kilocalories", "kilojoules"}), "/units"),
				),
				Entry("units calories; value missing",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("calories")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units calories; value out of range (lower)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("calories")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10000000.0), "/value"),
				),
				Entry("units calories; value in range (lower)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("calories")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units calories; value in range (upper)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("calories")
						datum.Value = pointer.FromFloat64(10000000.0)
					},
				),
				Entry("units calories; value out of range (upper)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("calories")
						datum.Value = pointer.FromFloat64(10000000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10000000.1, 0.0, 10000000.0), "/value"),
				),
				Entry("units joules; value missing",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("joules")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units joules; value out of range (lower)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("joules")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 41858000.0), "/value"),
				),
				Entry("units joules; value in range (lower)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("joules")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units joules; value in range (upper)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("joules")
						datum.Value = pointer.FromFloat64(41858000.0)
					},
				),
				Entry("units joules; value out of range (upper)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("joules")
						datum.Value = pointer.FromFloat64(41858000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(41858000.1, 0.0, 41858000.0), "/value"),
				),
				Entry("units kilocalories; value missing",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("kilocalories")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units kilocalories; value out of range (lower)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("kilocalories")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10000.0), "/value"),
				),
				Entry("units kilocalories; value in range (lower)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("kilocalories")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units kilocalories; value in range (upper)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("kilocalories")
						datum.Value = pointer.FromFloat64(10000.0)
					},
				),
				Entry("units kilocalories; value out of range (upper)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("kilocalories")
						datum.Value = pointer.FromFloat64(10000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10000.1, 0.0, 10000.0), "/value"),
				),
				Entry("units kilojoules; value missing",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("kilojoules")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units kilojoules; value out of range (lower)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("kilojoules")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 41858.0), "/value"),
				),
				Entry("units kilojoules; value in range (lower)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("kilojoules")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units kilojoules; value in range (upper)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("kilojoules")
						datum.Value = pointer.FromFloat64(41858.0)
					},
				),
				Entry("units kilojoules; value out of range (upper)",
					func(datum *food.Energy) {
						datum.Units = pointer.FromString("kilojoules")
						datum.Value = pointer.FromFloat64(41858.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(41858.1, 0.0, 41858.0), "/value"),
				),
				Entry("multiple errors",
					func(datum *food.Energy) {
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
				func(mutator func(datum *food.Energy)) {
					for _, origin := range structure.Origins() {
						datum := NewEnergy()
						mutator(datum)
						expectedDatum := CloneEnergy(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *food.Energy) {},
				),
				Entry("does not modify the datum; units missing",
					func(datum *food.Energy) { datum.Units = nil },
				),
				Entry("does not modify the datum; value missing",
					func(datum *food.Energy) { datum.Value = nil },
				),
			)
		})
	})

	Context("EnergyValueRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := food.EnergyValueRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := food.EnergyValueRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units calories", func() {
			minimum, maximum := food.EnergyValueRangeForUnits(pointer.FromString("calories"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(10000000.0))
		})

		It("returns expected range for units joules", func() {
			minimum, maximum := food.EnergyValueRangeForUnits(pointer.FromString("joules"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(41858000.0))
		})

		It("returns expected range for units kilocalories", func() {
			minimum, maximum := food.EnergyValueRangeForUnits(pointer.FromString("kilocalories"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(10000.0))
		})

		It("returns expected range for units kilojoules", func() {
			minimum, maximum := food.EnergyValueRangeForUnits(pointer.FromString("kilojoules"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(41858.0))
		})
	})
})
