package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"math"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewInsulin() *pump.Insulin {
	units := pointer.String(test.RandomStringFromArray(pump.InsulinUnits()))
	datum := pump.NewInsulin()
	datum.Duration = pointer.Float64(test.RandomFloat64FromRange(pump.InsulinDurationRangeForUnits(units)))
	datum.Units = units
	return datum
}

func CloneInsulin(datum *pump.Insulin) *pump.Insulin {
	if datum == nil {
		return nil
	}
	clone := pump.NewInsulin()
	clone.Duration = test.CloneFloat64(datum.Duration)
	clone.Units = test.CloneString(datum.Units)
	return clone
}

var _ = Describe("Insulin", func() {
	It("InsulinDurationHoursMaximum is expected", func() {
		Expect(pump.InsulinDurationHoursMaximum).To(Equal(10.0))
	})

	It("InsulinDurationHoursMinimum is expected", func() {
		Expect(pump.InsulinDurationHoursMinimum).To(Equal(0.0))
	})

	It("InsulinUnitsHours is expected", func() {
		Expect(pump.InsulinUnitsHours).To(Equal("hours"))
	})

	It("InsulinUnits returns expected", func() {
		Expect(pump.InsulinUnits()).To(Equal([]string{"hours"}))
	})

	Context("ParseInsulin", func() {
		// TODO
	})

	Context("NewInsulin", func() {
		It("is successful", func() {
			Expect(pump.NewInsulin()).To(Equal(&pump.Insulin{}))
		})
	})

	Context("Insulin", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.Insulin), expectedErrors ...error) {
					datum := NewInsulin()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.Insulin) {},
				),
				Entry("units missing; duration missing",
					func(datum *pump.Insulin) {
						datum.Duration = nil
						datum.Units = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/duration"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; duration out of range (lower)",
					func(datum *pump.Insulin) {
						datum.Duration = pointer.Float64(-0.1)
						datum.Units = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; duration in range (lower)",
					func(datum *pump.Insulin) {
						datum.Duration = pointer.Float64(0.0)
						datum.Units = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; duration in range (upper)",
					func(datum *pump.Insulin) {
						datum.Duration = pointer.Float64(10.0)
						datum.Units = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; duration out of range (upper)",
					func(datum *pump.Insulin) {
						datum.Duration = pointer.Float64(10.1)
						datum.Units = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; duration missing",
					func(datum *pump.Insulin) {
						datum.Duration = nil
						datum.Units = pointer.String("invalid")
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/duration"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours"}), "/units"),
				),
				Entry("units invalid; duration out of range (lower)",
					func(datum *pump.Insulin) {
						datum.Duration = pointer.Float64(-0.1)
						datum.Units = pointer.String("invalid")
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours"}), "/units"),
				),
				Entry("units invalid; duration in range (lower)",
					func(datum *pump.Insulin) {
						datum.Duration = pointer.Float64(0.0)
						datum.Units = pointer.String("invalid")
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours"}), "/units"),
				),
				Entry("units invalid; duration in range (upper)",
					func(datum *pump.Insulin) {
						datum.Duration = pointer.Float64(10.0)
						datum.Units = pointer.String("invalid")
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours"}), "/units"),
				),
				Entry("units invalid; duration out of range (upper)",
					func(datum *pump.Insulin) {
						datum.Duration = pointer.Float64(10.1)
						datum.Units = pointer.String("invalid")
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours"}), "/units"),
				),
				Entry("units hours: duration missing",
					func(datum *pump.Insulin) {
						datum.Duration = nil
						datum.Units = pointer.String("hours")
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/duration"),
				),
				Entry("units hours: duration out of range (lower)",
					func(datum *pump.Insulin) {
						datum.Duration = pointer.Float64(-0.1)
						datum.Units = pointer.String("hours")
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10.0), "/duration"),
				),
				Entry("units hours: duration in range (lower)",
					func(datum *pump.Insulin) {
						datum.Duration = pointer.Float64(0.0)
						datum.Units = pointer.String("hours")
					},
				),
				Entry("units hours: duration in range (upper)",
					func(datum *pump.Insulin) {
						datum.Duration = pointer.Float64(10.0)
						datum.Units = pointer.String("hours")
					},
				),
				Entry("units hours: duration out of range (upper)",
					func(datum *pump.Insulin) {
						datum.Duration = pointer.Float64(10.1)
						datum.Units = pointer.String("hours")
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(10.1, 0.0, 10.0), "/duration"),
				),
				Entry("units missing",
					func(datum *pump.Insulin) { datum.Units = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *pump.Insulin) { datum.Units = pointer.String("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours"}), "/units"),
				),
				Entry("units hours",
					func(datum *pump.Insulin) {
						datum.Duration = pointer.Float64(0.0)
						datum.Units = pointer.String("hours")
					},
				),
				Entry("multiple errors",
					func(datum *pump.Insulin) {
						datum.Duration = nil
						datum.Units = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/duration"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.Insulin)) {
					for _, origin := range structure.Origins() {
						datum := NewInsulin()
						mutator(datum)
						expectedDatum := CloneInsulin(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.Insulin) {},
				),
				Entry("does not modify the datum; duration missing",
					func(datum *pump.Insulin) { datum.Duration = nil },
				),
				Entry("does not modify the datum; units missing",
					func(datum *pump.Insulin) { datum.Units = nil },
				),
			)
		})
	})

	Context("InsulinDurationRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := pump.InsulinDurationRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := pump.InsulinDurationRangeForUnits(pointer.String("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units units", func() {
			minimum, maximum := pump.InsulinDurationRangeForUnits(pointer.String("hours"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(10.0))
		})
	})
})
