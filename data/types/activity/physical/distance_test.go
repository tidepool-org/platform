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

func NewDistance() *physical.Distance {
	datum := physical.NewDistance()
	datum.Units = pointer.FromString(test.RandomStringFromArray(physical.DistanceUnits()))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(physical.DistanceValueRangeForUnits(datum.Units)))
	return datum
}

func CloneDistance(datum *physical.Distance) *physical.Distance {
	if datum == nil {
		return nil
	}
	clone := physical.NewDistance()
	clone.Units = test.CloneString(datum.Units)
	clone.Value = test.CloneFloat64(datum.Value)
	return clone
}

var _ = Describe("Distance", func() {
	It("DistanceFeetPerMile is expected", func() {
		Expect(physical.DistanceFeetPerMile).To(Equal(5280.0))
	})

	It("DistanceKilometersPerMile is expected", func() {
		Expect(physical.DistanceKilometersPerMile).To(Equal(1.609344))
	})

	It("DistanceMetersPerMile is expected", func() {
		Expect(physical.DistanceMetersPerMile).To(Equal(1609.344))
	})

	It("DistanceUnitsFeet is expected", func() {
		Expect(physical.DistanceUnitsFeet).To(Equal("feet"))
	})

	It("DistanceUnitsKilometers is expected", func() {
		Expect(physical.DistanceUnitsKilometers).To(Equal("kilometers"))
	})

	It("DistanceUnitsMeters is expected", func() {
		Expect(physical.DistanceUnitsMeters).To(Equal("meters"))
	})

	It("DistanceUnitsMiles is expected", func() {
		Expect(physical.DistanceUnitsMiles).To(Equal("miles"))
	})

	It("DistanceUnitsYards is expected", func() {
		Expect(physical.DistanceUnitsYards).To(Equal("yards"))
	})

	It("DistanceValueFeetMaximum is expected", func() {
		Expect(physical.DistanceValueFeetMaximum).To(Equal(528000.0))
	})

	It("DistanceValueFeetMinimum is expected", func() {
		Expect(physical.DistanceValueFeetMinimum).To(Equal(0.0))
	})

	It("DistanceValueKilometersMaximum is expected", func() {
		Expect(physical.DistanceValueKilometersMaximum).To(Equal(160.9344))
	})

	It("DistanceValueKilometersMinimum is expected", func() {
		Expect(physical.DistanceValueKilometersMinimum).To(Equal(0.0))
	})

	It("DistanceValueMetersMaximum is expected", func() {
		Expect(physical.DistanceValueMetersMaximum).To(Equal(160934.4))
	})

	It("DistanceValueMetersMinimum is expected", func() {
		Expect(physical.DistanceValueMetersMinimum).To(Equal(0.0))
	})

	It("DistanceValueMilesMaximum is expected", func() {
		Expect(physical.DistanceValueMilesMaximum).To(Equal(100.0))
	})

	It("DistanceValueMilesMinimum is expected", func() {
		Expect(physical.DistanceValueMilesMinimum).To(Equal(0.0))
	})

	It("DistanceValueYardsMaximum is expected", func() {
		Expect(physical.DistanceValueYardsMaximum).To(Equal(176000.0))
	})

	It("DistanceValueYardsMinimum is expected", func() {
		Expect(physical.DistanceValueYardsMinimum).To(Equal(0.0))
	})

	It("DistanceUnits returns expected", func() {
		Expect(physical.DistanceUnits()).To(Equal([]string{"feet", "kilometers", "meters", "miles", "yards"}))
	})

	Context("ParseDistance", func() {
		// TODO
	})

	Context("NewDistance", func() {
		It("returns the expected datum", func() {
			Expect(physical.NewDistance()).To(Equal(&physical.Distance{}))
		})
	})

	Context("Distance", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *physical.Distance), expectedErrors ...error) {
					datum := NewDistance()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *physical.Distance) {},
				),
				Entry("units missing; value missing",
					func(datum *physical.Distance) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value out of range (lower)",
					func(datum *physical.Distance) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *physical.Distance) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *physical.Distance) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(528000.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value out of range (upper)",
					func(datum *physical.Distance) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(528000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "kilometers", "meters", "miles", "yards"}), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value out of range (lower)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "kilometers", "meters", "miles", "yards"}), "/units"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "kilometers", "meters", "miles", "yards"}), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(528000.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "kilometers", "meters", "miles", "yards"}), "/units"),
				),
				Entry("units invalid; value out of range (upper)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(528000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "kilometers", "meters", "miles", "yards"}), "/units"),
				),
				Entry("units feet; value missing",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("feet")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units feet; value out of range (lower)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("feet")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 528000.0), "/value"),
				),
				Entry("units feet; value in range (lower)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("feet")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units feet; value in range (upper)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("feet")
						datum.Value = pointer.FromFloat64(528000.0)
					},
				),
				Entry("units feet; value out of range (upper)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("feet")
						datum.Value = pointer.FromFloat64(528000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(528000.1, 0.0, 528000.0), "/value"),
				),
				Entry("units kilometers; value missing",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("kilometers")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units kilometers; value out of range (lower)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("kilometers")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 160.9344), "/value"),
				),
				Entry("units kilometers; value in range (lower)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("kilometers")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units kilometers; value in range (upper)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("kilometers")
						datum.Value = pointer.FromFloat64(160.9344)
					},
				),
				Entry("units kilometers; value out of range (upper)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("kilometers")
						datum.Value = pointer.FromFloat64(160.9345)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(160.9345, 0.0, 160.9344), "/value"),
				),
				Entry("units meters; value missing",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("meters")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units meters; value out of range (lower)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("meters")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 160934.4), "/value"),
				),
				Entry("units meters; value in range (lower)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("meters")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units meters; value in range (upper)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("meters")
						datum.Value = pointer.FromFloat64(160934.4)
					},
				),
				Entry("units meters; value out of range (upper)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("meters")
						datum.Value = pointer.FromFloat64(160934.5)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(160934.5, 0.0, 160934.4), "/value"),
				),
				Entry("units miles; value missing",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("miles")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units miles; value out of range (lower)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("miles")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/value"),
				),
				Entry("units miles; value in range (lower)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("miles")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units miles; value in range (upper)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("miles")
						datum.Value = pointer.FromFloat64(100.0)
					},
				),
				Entry("units miles; value out of range (upper)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("miles")
						datum.Value = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/value"),
				),
				Entry("units yards; value missing",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("yards")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units yards; value out of range (lower)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("yards")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 176000.0), "/value"),
				),
				Entry("units yards; value in range (lower)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("yards")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units yards; value in range (upper)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("yards")
						datum.Value = pointer.FromFloat64(176000.0)
					},
				),
				Entry("units yards; value out of range (upper)",
					func(datum *physical.Distance) {
						datum.Units = pointer.FromString("yards")
						datum.Value = pointer.FromFloat64(176000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(176000.1, 0.0, 176000.0), "/value"),
				),
				Entry("multiple errors",
					func(datum *physical.Distance) {
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
				func(mutator func(datum *physical.Distance)) {
					for _, origin := range structure.Origins() {
						datum := NewDistance()
						mutator(datum)
						expectedDatum := CloneDistance(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *physical.Distance) {},
				),
				Entry("does not modify the datum; units missing",
					func(datum *physical.Distance) { datum.Units = nil },
				),
				Entry("does not modify the datum; units feet",
					func(datum *physical.Distance) { datum.Units = pointer.FromString("feet") },
				),
				Entry("does not modify the datum; units kilometers",
					func(datum *physical.Distance) { datum.Units = pointer.FromString("kilometers") },
				),
				Entry("does not modify the datum; units meters",
					func(datum *physical.Distance) { datum.Units = pointer.FromString("meters") },
				),
				Entry("does not modify the datum; units miles",
					func(datum *physical.Distance) { datum.Units = pointer.FromString("miles") },
				),
				Entry("does not modify the datum; units yards",
					func(datum *physical.Distance) { datum.Units = pointer.FromString("yards") },
				),
				Entry("does not modify the datum; value missing",
					func(datum *physical.Distance) { datum.Value = nil },
				),
			)
		})
	})

	Context("DistanceValueRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := physical.DistanceValueRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := physical.DistanceValueRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units feet", func() {
			minimum, maximum := physical.DistanceValueRangeForUnits(pointer.FromString("feet"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(528000.0))
		})

		It("returns expected range for units kilometers", func() {
			minimum, maximum := physical.DistanceValueRangeForUnits(pointer.FromString("kilometers"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(160.9344))
		})

		It("returns expected range for units meters", func() {
			minimum, maximum := physical.DistanceValueRangeForUnits(pointer.FromString("meters"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(160934.4))
		})

		It("returns expected range for units miles", func() {
			minimum, maximum := physical.DistanceValueRangeForUnits(pointer.FromString("miles"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(100.0))
		})

		It("returns expected range for units yards", func() {
			minimum, maximum := physical.DistanceValueRangeForUnits(pointer.FromString("yards"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(176000.0))
		})
	})
})
