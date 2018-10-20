package location_test

import (
	"math"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/common/location"
	testDataTypesCommonLocation "github.com/tidepool-org/platform/data/types/common/location/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Elevation", func() {
	It("ElevationUnitsFeet is expected", func() {
		Expect(location.ElevationUnitsFeet).To(Equal("feet"))
	})

	It("ElevationUnitsMeter is expected", func() {
		Expect(location.ElevationUnitsMeter).To(Equal("meters"))
	})

	It("ElevationValueFeetMaximum is expected", func() {
		Expect(location.ElevationValueFeetMaximum).To(Equal(10000.0 / 0.3048))
	})

	It("ElevationValueFeetMinimum is expected", func() {
		Expect(location.ElevationValueFeetMinimum).To(Equal(0.0))
	})

	It("ElevationValueMetersMaximum is expected", func() {
		Expect(location.ElevationValueMetersMaximum).To(Equal(10000.0))
	})

	It("ElevationValueMetersMinimum is expected", func() {
		Expect(location.ElevationValueMetersMinimum).To(Equal(0.0))
	})

	It("ElevationUnits returns expected", func() {
		Expect(location.ElevationUnits()).To(Equal([]string{"feet", "meters"}))
	})

	Context("ParseElevation", func() {
		// TODO
	})

	Context("NewElevation", func() {
		It("is successful", func() {
			Expect(location.NewElevation()).To(Equal(&location.Elevation{}))
		})
	})

	Context("Elevation", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *location.Elevation, units *string), units *string, expectedErrors ...error) {
					datum := testDataTypesCommonLocation.NewElevation(units)
					mutator(datum, units)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *location.Elevation, units *string) {},
					pointer.FromString("feet"),
				),
				Entry("units missing",
					func(datum *location.Elevation, units *string) {},
					nil,
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *location.Elevation, units *string) {},
					pointer.FromString("invalid"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
				),
				Entry("units feet",
					func(datum *location.Elevation, units *string) {},
					pointer.FromString("feet"),
				),
				Entry("units meters",
					func(datum *location.Elevation, units *string) {},
					pointer.FromString("meters"),
				),
				Entry("units missing; value missing",
					func(datum *location.Elevation, units *string) { datum.Value = nil },
					nil,
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *location.Elevation, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					nil,
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *location.Elevation, units *string) { datum.Value = pointer.FromFloat64(3281.9) },
					nil,
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *location.Elevation, units *string) { datum.Value = nil },
					pointer.FromString("invalid"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *location.Elevation, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					pointer.FromString("invalid"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *location.Elevation, units *string) { datum.Value = pointer.FromFloat64(3281.9) },
					pointer.FromString("invalid"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
				),
				Entry("units feet; value missing",
					func(datum *location.Elevation, units *string) { datum.Value = nil },
					pointer.FromString("feet"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units feet; value out of range (lower)",
					func(datum *location.Elevation, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					pointer.FromString("feet"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10000.0/0.3048), "/value"),
				),
				Entry("units feet; value in range (lower)",
					func(datum *location.Elevation, units *string) { datum.Value = pointer.FromFloat64(0.0) },
					pointer.FromString("feet"),
				),
				Entry("units feet; value in range (upper)",
					func(datum *location.Elevation, units *string) { datum.Value = pointer.FromFloat64(32808.3) },
					pointer.FromString("feet"),
				),
				Entry("units feet; value out of range (upper)",
					func(datum *location.Elevation, units *string) { datum.Value = pointer.FromFloat64(32808.4) },
					pointer.FromString("feet"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(32808.4, 0.0, 10000.0/0.3048), "/value"),
				),
				Entry("units meters; value missing",
					func(datum *location.Elevation, units *string) { datum.Value = nil },
					pointer.FromString("meters"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units meters; value out of range (lower)",
					func(datum *location.Elevation, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					pointer.FromString("meters"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10000.0), "/value"),
				),
				Entry("units meters; value in range (lower)",
					func(datum *location.Elevation, units *string) { datum.Value = pointer.FromFloat64(0.0) },
					pointer.FromString("meters"),
				),
				Entry("units meters; value in range (upper)",
					func(datum *location.Elevation, units *string) { datum.Value = pointer.FromFloat64(10000.0) },
					pointer.FromString("meters"),
				),
				Entry("units meters; value out of range (upper)",
					func(datum *location.Elevation, units *string) { datum.Value = pointer.FromFloat64(10000.1) },
					pointer.FromString("meters"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(10000.1, 0.0, 10000.0), "/value"),
				),
				Entry("multiple errors",
					func(datum *location.Elevation, units *string) {
						datum.Value = nil
					},
					nil,
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *location.Elevation, units *string), units *string) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesCommonLocation.NewElevation(units)
						mutator(datum, units)
						expectedDatum := testDataTypesCommonLocation.CloneElevation(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *location.Elevation, units *string) {},
					pointer.FromString("feet"),
				),
				Entry("does not modify the datum; units missing",
					func(datum *location.Elevation, units *string) { datum.Units = nil },
					nil,
				),
				Entry("does not modify the datum; value missing",
					func(datum *location.Elevation, units *string) { datum.Value = nil },
					pointer.FromString("feet"),
				),
				Entry("does not modify the datum; all missing",
					func(datum *location.Elevation, units *string) { *datum = location.Elevation{} },
					pointer.FromString("feet"),
				),
			)
		})
	})

	Context("ElevationValueRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := location.ElevationValueRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := location.ElevationValueRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units feet", func() {
			minimum, maximum := location.ElevationValueRangeForUnits(pointer.FromString("feet"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(10000.0 / 0.3048))
		})

		It("returns expected range for units meters", func() {
			minimum, maximum := location.ElevationValueRangeForUnits(pointer.FromString("meters"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(10000.0))
		})
	})
})
