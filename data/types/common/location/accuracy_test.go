package location_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"math"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/common/location"
	testDataTypesCommonLocation "github.com/tidepool-org/platform/data/types/common/location/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Accuracy", func() {
	It("AccuracyUnitsFeet is expected", func() {
		Expect(location.AccuracyUnitsFeet).To(Equal("feet"))
	})

	It("AccuracyUnitsMeter is expected", func() {
		Expect(location.AccuracyUnitsMeter).To(Equal("meters"))
	})

	It("AccuracyValueFeetMaximum is expected", func() {
		Expect(location.AccuracyValueFeetMaximum).To(Equal(1000.0 / 0.3048))
	})

	It("AccuracyValueFeetMinimum is expected", func() {
		Expect(location.AccuracyValueFeetMinimum).To(Equal(0.0))
	})

	It("AccuracyValueMetersMaximum is expected", func() {
		Expect(location.AccuracyValueMetersMaximum).To(Equal(1000.0))
	})

	It("AccuracyValueMetersMinimum is expected", func() {
		Expect(location.AccuracyValueMetersMinimum).To(Equal(0.0))
	})

	It("AccuracyUnits returns expected", func() {
		Expect(location.AccuracyUnits()).To(Equal([]string{"feet", "meters"}))
	})

	Context("ParseAccuracy", func() {
		// TODO
	})

	Context("NewAccuracy", func() {
		It("is successful", func() {
			Expect(location.NewAccuracy()).To(Equal(&location.Accuracy{}))
		})
	})

	Context("Accuracy", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *location.Accuracy, units *string), units *string, expectedErrors ...error) {
					datum := testDataTypesCommonLocation.NewAccuracy(units)
					mutator(datum, units)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *location.Accuracy, units *string) {},
					pointer.FromString("feet"),
				),
				Entry("units missing",
					func(datum *location.Accuracy, units *string) {},
					nil,
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *location.Accuracy, units *string) {},
					pointer.FromString("invalid"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
				),
				Entry("units feet",
					func(datum *location.Accuracy, units *string) {},
					pointer.FromString("feet"),
				),
				Entry("units meters",
					func(datum *location.Accuracy, units *string) {},
					pointer.FromString("meters"),
				),
				Entry("units missing; value missing",
					func(datum *location.Accuracy, units *string) { datum.Value = nil },
					nil,
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *location.Accuracy, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					nil,
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *location.Accuracy, units *string) { datum.Value = pointer.FromFloat64(3281.9) },
					nil,
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *location.Accuracy, units *string) { datum.Value = nil },
					pointer.FromString("invalid"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *location.Accuracy, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					pointer.FromString("invalid"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *location.Accuracy, units *string) { datum.Value = pointer.FromFloat64(3281.9) },
					pointer.FromString("invalid"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
				),
				Entry("units feet; value missing",
					func(datum *location.Accuracy, units *string) { datum.Value = nil },
					pointer.FromString("feet"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units feet; value out of range (lower)",
					func(datum *location.Accuracy, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					pointer.FromString("feet"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0/0.3048), "/value"),
				),
				Entry("units feet; value in range (lower)",
					func(datum *location.Accuracy, units *string) { datum.Value = pointer.FromFloat64(0.0) },
					pointer.FromString("feet"),
				),
				Entry("units feet; value in range (upper)",
					func(datum *location.Accuracy, units *string) { datum.Value = pointer.FromFloat64(3280.8) },
					pointer.FromString("feet"),
				),
				Entry("units feet; value out of range (upper)",
					func(datum *location.Accuracy, units *string) { datum.Value = pointer.FromFloat64(3281.9) },
					pointer.FromString("feet"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(3281.9, 0.0, 1000.0/0.3048), "/value"),
				),
				Entry("units meters; value missing",
					func(datum *location.Accuracy, units *string) { datum.Value = nil },
					pointer.FromString("meters"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units meters; value out of range (lower)",
					func(datum *location.Accuracy, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					pointer.FromString("meters"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value"),
				),
				Entry("units meters; value in range (lower)",
					func(datum *location.Accuracy, units *string) { datum.Value = pointer.FromFloat64(0.0) },
					pointer.FromString("meters"),
				),
				Entry("units meters; value in range (upper)",
					func(datum *location.Accuracy, units *string) { datum.Value = pointer.FromFloat64(1000.0) },
					pointer.FromString("meters"),
				),
				Entry("units meters; value out of range (upper)",
					func(datum *location.Accuracy, units *string) { datum.Value = pointer.FromFloat64(1000.1) },
					pointer.FromString("meters"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value"),
				),
				Entry("multiple errors",
					func(datum *location.Accuracy, units *string) {
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
				func(mutator func(datum *location.Accuracy, units *string), units *string) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesCommonLocation.NewAccuracy(units)
						mutator(datum, units)
						expectedDatum := testDataTypesCommonLocation.CloneAccuracy(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *location.Accuracy, units *string) {},
					pointer.FromString("feet"),
				),
				Entry("does not modify the datum; units missing",
					func(datum *location.Accuracy, units *string) { datum.Units = nil },
					nil,
				),
				Entry("does not modify the datum; value missing",
					func(datum *location.Accuracy, units *string) { datum.Value = nil },
					pointer.FromString("feet"),
				),
				Entry("does not modify the datum; all missing",
					func(datum *location.Accuracy, units *string) { *datum = location.Accuracy{} },
					pointer.FromString("feet"),
				),
			)
		})
	})

	Context("AccuracyValueRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := location.AccuracyValueRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := location.AccuracyValueRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units feet", func() {
			minimum, maximum := location.AccuracyValueRangeForUnits(pointer.FromString("feet"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(1000.0 / 0.3048))
		})

		It("returns expected range for units meters", func() {
			minimum, maximum := location.AccuracyValueRangeForUnits(pointer.FromString("meters"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(1000.0))
		})
	})
})
