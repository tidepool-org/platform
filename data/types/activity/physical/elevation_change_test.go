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

func NewElevationChange() *physical.ElevationChange {
	datum := physical.NewElevationChange()
	datum.Units = pointer.FromString(test.RandomStringFromArray(physical.ElevationChangeUnits()))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(physical.ElevationChangeValueRangeForUnits(datum.Units)))
	return datum
}

func CloneElevationChange(datum *physical.ElevationChange) *physical.ElevationChange {
	if datum == nil {
		return nil
	}
	clone := physical.NewElevationChange()
	clone.Units = test.CloneString(datum.Units)
	clone.Value = test.CloneFloat64(datum.Value)
	return clone
}

var _ = Describe("ElevationChange", func() {
	It("ElevationChangeMetersPerFoot is expected", func() {
		Expect(physical.ElevationChangeMetersPerFoot).To(Equal(0.3048))
	})

	It("ElevationChangeUnitsFeet is expected", func() {
		Expect(physical.ElevationChangeUnitsFeet).To(Equal("feet"))
	})

	It("ElevationChangeUnitsMeters is expected", func() {
		Expect(physical.ElevationChangeUnitsMeters).To(Equal("meters"))
	})

	It("ElevationChangeValueFeetMaximum is expected", func() {
		Expect(physical.ElevationChangeValueFeetMaximum).To(Equal(52800.0))
	})

	It("ElevationChangeValueFeetMinimum is expected", func() {
		Expect(physical.ElevationChangeValueFeetMinimum).To(Equal(0.0))
	})

	It("ElevationChangeValueMetersMaximum is expected", func() {
		Expect(physical.ElevationChangeValueMetersMaximum).To(Equal(16093.44))
	})

	It("ElevationChangeValueMetersMinimum is expected", func() {
		Expect(physical.ElevationChangeValueMetersMinimum).To(Equal(0.0))
	})

	It("ElevationChangeUnits returns expected", func() {
		Expect(physical.ElevationChangeUnits()).To(Equal([]string{"feet", "meters"}))
	})

	Context("ParseElevationChange", func() {
		// TODO
	})

	Context("NewElevationChange", func() {
		It("returns the expected datum", func() {
			Expect(physical.NewElevationChange()).To(Equal(&physical.ElevationChange{}))
		})
	})

	Context("ElevationChange", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *physical.ElevationChange), expectedErrors ...error) {
					datum := NewElevationChange()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *physical.ElevationChange) {},
				),
				Entry("units missing; value missing",
					func(datum *physical.ElevationChange) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value out of range (lower)",
					func(datum *physical.ElevationChange) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *physical.ElevationChange) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *physical.ElevationChange) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(528000.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value out of range (upper)",
					func(datum *physical.ElevationChange) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(528000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *physical.ElevationChange) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value out of range (lower)",
					func(datum *physical.ElevationChange) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *physical.ElevationChange) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *physical.ElevationChange) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(528000.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
				),
				Entry("units invalid; value out of range (upper)",
					func(datum *physical.ElevationChange) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(528000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
				),
				Entry("units feet; value missing",
					func(datum *physical.ElevationChange) {
						datum.Units = pointer.FromString("feet")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units feet; value out of range (lower)",
					func(datum *physical.ElevationChange) {
						datum.Units = pointer.FromString("feet")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 52800.0), "/value"),
				),
				Entry("units feet; value in range (lower)",
					func(datum *physical.ElevationChange) {
						datum.Units = pointer.FromString("feet")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units feet; value in range (upper)",
					func(datum *physical.ElevationChange) {
						datum.Units = pointer.FromString("feet")
						datum.Value = pointer.FromFloat64(52800.0)
					},
				),
				Entry("units feet; value out of range (upper)",
					func(datum *physical.ElevationChange) {
						datum.Units = pointer.FromString("feet")
						datum.Value = pointer.FromFloat64(52800.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(52800.1, 0.0, 52800.0), "/value"),
				),
				Entry("units meters; value missing",
					func(datum *physical.ElevationChange) {
						datum.Units = pointer.FromString("meters")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units meters; value out of range (lower)",
					func(datum *physical.ElevationChange) {
						datum.Units = pointer.FromString("meters")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 16093.44), "/value"),
				),
				Entry("units meters; value in range (lower)",
					func(datum *physical.ElevationChange) {
						datum.Units = pointer.FromString("meters")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units meters; value in range (upper)",
					func(datum *physical.ElevationChange) {
						datum.Units = pointer.FromString("meters")
						datum.Value = pointer.FromFloat64(16093.44)
					},
				),
				Entry("units meters; value out of range (upper)",
					func(datum *physical.ElevationChange) {
						datum.Units = pointer.FromString("meters")
						datum.Value = pointer.FromFloat64(16093.45)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(16093.45, 0.0, 16093.44), "/value"),
				),
				Entry("multiple errors",
					func(datum *physical.ElevationChange) {
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
				func(mutator func(datum *physical.ElevationChange)) {
					for _, origin := range structure.Origins() {
						datum := NewElevationChange()
						mutator(datum)
						expectedDatum := CloneElevationChange(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *physical.ElevationChange) {},
				),
				Entry("does not modify the datum; units missing",
					func(datum *physical.ElevationChange) { datum.Units = nil },
				),
				Entry("does not modify the datum; units feet",
					func(datum *physical.ElevationChange) { datum.Units = pointer.FromString("feet") },
				),
				Entry("does not modify the datum; units meters",
					func(datum *physical.ElevationChange) { datum.Units = pointer.FromString("meters") },
				),
				Entry("does not modify the datum; value missing",
					func(datum *physical.ElevationChange) { datum.Value = nil },
				),
			)
		})
	})

	Context("ElevationChangeValueRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := physical.ElevationChangeValueRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := physical.ElevationChangeValueRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units feet", func() {
			minimum, maximum := physical.ElevationChangeValueRangeForUnits(pointer.FromString("feet"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(52800.0))
		})

		It("returns expected range for units meters", func() {
			minimum, maximum := physical.ElevationChangeValueRangeForUnits(pointer.FromString("meters"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(16093.44))
		})
	})
})
