package physical_test

import (
	"math"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/activity/physical"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewEnergy() *physical.Energy {
	datum := physical.NewEnergy()
	datum.Units = pointer.FromString(test.RandomStringFromArray(physical.EnergyUnits()))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(physical.EnergyValueRangeForUnits(datum.Units)))
	return datum
}

func CloneEnergy(datum *physical.Energy) *physical.Energy {
	if datum == nil {
		return nil
	}
	clone := physical.NewEnergy()
	clone.Units = test.CloneString(datum.Units)
	clone.Value = test.CloneFloat64(datum.Value)
	return clone
}

var _ = Describe("Energy", func() {
	It("EnergyKilojoulesPerKilocalorie is expected", func() {
		Expect(physical.EnergyKilojoulesPerKilocalorie).To(Equal(4.1858))
	})

	It("EnergyUnitsCalories is expected", func() {
		Expect(physical.EnergyUnitsCalories).To(Equal("calories"))
	})

	It("EnergyUnitsJoules is expected", func() {
		Expect(physical.EnergyUnitsJoules).To(Equal("joules"))
	})

	It("EnergyUnitsKilocalories is expected", func() {
		Expect(physical.EnergyUnitsKilocalories).To(Equal("kilocalories"))
	})

	It("EnergyUnitsKilojoules is expected", func() {
		Expect(physical.EnergyUnitsKilojoules).To(Equal("kilojoules"))
	})

	It("EnergyValueCaloriesMaximum is expected", func() {
		Expect(physical.EnergyValueCaloriesMaximum).To(Equal(10000000.0))
	})

	It("EnergyValueCaloriesMinimum is expected", func() {
		Expect(physical.EnergyValueCaloriesMinimum).To(Equal(0.0))
	})

	It("EnergyValueJoulesMaximum is expected", func() {
		Expect(physical.EnergyValueJoulesMaximum).To(Equal(41858000.0))
	})

	It("EnergyValueJoulesMinimum is expected", func() {
		Expect(physical.EnergyValueJoulesMinimum).To(Equal(0.0))
	})

	It("EnergyValueKilocaloriesMaximum is expected", func() {
		Expect(physical.EnergyValueKilocaloriesMaximum).To(Equal(10000.0))
	})

	It("EnergyValueKilocaloriesMinimum is expected", func() {
		Expect(physical.EnergyValueKilocaloriesMinimum).To(Equal(0.0))
	})

	It("EnergyValueKilojoulesMaximum is expected", func() {
		Expect(physical.EnergyValueKilojoulesMaximum).To(Equal(41858.0))
	})

	It("EnergyValueKilojoulesMinimum is expected", func() {
		Expect(physical.EnergyValueKilojoulesMinimum).To(Equal(0.0))
	})

	It("EnergyUnits returns expected", func() {
		Expect(physical.EnergyUnits()).To(Equal([]string{"calories", "joules", "kilocalories", "kilojoules"}))
	})

	Context("ParseEnergy", func() {
		// TODO
	})

	Context("NewEnergy", func() {
		It("is successful", func() {
			Expect(physical.NewEnergy()).To(Equal(&physical.Energy{}))
		})
	})

	Context("Energy", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *physical.Energy), expectedErrors ...error) {
					datum := NewEnergy()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *physical.Energy) {},
				),
				Entry("units missing",
					func(datum *physical.Energy) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *physical.Energy) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"calories", "joules", "kilocalories", "kilojoules"}), "/units"),
				),
				Entry("units calories",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("calories")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units joules",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("joules")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units kilocalories",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("kilocalories")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units kilojoules",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("kilojoules")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units missing; value missing",
					func(datum *physical.Energy) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value out of range (lower)",
					func(datum *physical.Energy) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *physical.Energy) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *physical.Energy) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(41858000.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value out of range (upper)",
					func(datum *physical.Energy) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(41858000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"calories", "joules", "kilocalories", "kilojoules"}), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value out of range (lower)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"calories", "joules", "kilocalories", "kilojoules"}), "/units"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"calories", "joules", "kilocalories", "kilojoules"}), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(41858000.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"calories", "joules", "kilocalories", "kilojoules"}), "/units"),
				),
				Entry("units invalid; value out of range (upper)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(41858000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"calories", "joules", "kilocalories", "kilojoules"}), "/units"),
				),
				Entry("units calories; value missing",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("calories")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units calories; value out of range (lower)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("calories")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10000000.0), "/value"),
				),
				Entry("units calories; value in range (lower)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("calories")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units calories; value in range (upper)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("calories")
						datum.Value = pointer.FromFloat64(10000000.0)
					},
				),
				Entry("units calories; value out of range (upper)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("calories")
						datum.Value = pointer.FromFloat64(10000000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10000000.1, 0.0, 10000000.0), "/value"),
				),
				Entry("units joules; value missing",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("joules")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units joules; value out of range (lower)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("joules")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 41858000.0), "/value"),
				),
				Entry("units joules; value in range (lower)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("joules")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units joules; value in range (upper)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("joules")
						datum.Value = pointer.FromFloat64(41858000.0)
					},
				),
				Entry("units joules; value out of range (upper)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("joules")
						datum.Value = pointer.FromFloat64(41858000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(41858000.1, 0.0, 41858000.0), "/value"),
				),
				Entry("units kilocalories; value missing",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("kilocalories")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units kilocalories; value out of range (lower)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("kilocalories")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10000.0), "/value"),
				),
				Entry("units kilocalories; value in range (lower)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("kilocalories")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units kilocalories; value in range (upper)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("kilocalories")
						datum.Value = pointer.FromFloat64(10000.0)
					},
				),
				Entry("units kilocalories; value out of range (upper)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("kilocalories")
						datum.Value = pointer.FromFloat64(10000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10000.1, 0.0, 10000.0), "/value"),
				),
				Entry("units kilojoules; value missing",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("kilojoules")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units kilojoules; value out of range (lower)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("kilojoules")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 41858.0), "/value"),
				),
				Entry("units kilojoules; value in range (lower)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("kilojoules")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units kilojoules; value in range (upper)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("kilojoules")
						datum.Value = pointer.FromFloat64(41858.0)
					},
				),
				Entry("units kilojoules; value out of range (upper)",
					func(datum *physical.Energy) {
						datum.Units = pointer.FromString("kilojoules")
						datum.Value = pointer.FromFloat64(41858.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(41858.1, 0.0, 41858.0), "/value"),
				),
				Entry("multiple errors",
					func(datum *physical.Energy) {
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
				func(mutator func(datum *physical.Energy)) {
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
					func(datum *physical.Energy) {},
				),
				Entry("does not modify the datum; units missing",
					func(datum *physical.Energy) { datum.Units = nil },
				),
				Entry("does not modify the datum; value missing",
					func(datum *physical.Energy) { datum.Value = nil },
				),
			)
		})
	})

	Context("EnergyValueRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := physical.EnergyValueRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := physical.EnergyValueRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units calories", func() {
			minimum, maximum := physical.EnergyValueRangeForUnits(pointer.FromString("calories"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(10000000.0))
		})

		It("returns expected range for units joules", func() {
			minimum, maximum := physical.EnergyValueRangeForUnits(pointer.FromString("joules"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(41858000.0))
		})

		It("returns expected range for units kilocalories", func() {
			minimum, maximum := physical.EnergyValueRangeForUnits(pointer.FromString("kilocalories"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(10000.0))
		})

		It("returns expected range for units kilojoules", func() {
			minimum, maximum := physical.EnergyValueRangeForUnits(pointer.FromString("kilojoules"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(41858.0))
		})
	})
})
