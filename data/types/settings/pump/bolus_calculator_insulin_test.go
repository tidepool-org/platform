package pump_test

import (
	"math"

	pumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("BolusCalculatorInsulin", func() {
	It("BolusCalculatorInsulinDurationHoursMaximum is expected", func() {
		Expect(pump.BolusCalculatorInsulinDurationHoursMaximum).To(Equal(10.0))
	})

	It("BolusCalculatorInsulinDurationHoursMinimum is expected", func() {
		Expect(pump.BolusCalculatorInsulinDurationHoursMinimum).To(Equal(0.0))
	})

	It("BolusCalculatorInsulinDurationMinutesMaximum is expected", func() {
		Expect(pump.BolusCalculatorInsulinDurationMinutesMaximum).To(Equal(600.0))
	})

	It("BolusCalculatorInsulinDurationMinutesMinimum is expected", func() {
		Expect(pump.BolusCalculatorInsulinDurationMinutesMinimum).To(Equal(0.0))
	})

	It("BolusCalculatorInsulinDurationSecondsMaximum is expected", func() {
		Expect(pump.BolusCalculatorInsulinDurationSecondsMaximum).To(Equal(36000.0))
	})

	It("BolusCalculatorInsulinDurationSecondsMinimum is expected", func() {
		Expect(pump.BolusCalculatorInsulinDurationSecondsMinimum).To(Equal(0.0))
	})

	It("BolusCalculatorInsulinUnitsHours is expected", func() {
		Expect(pump.BolusCalculatorInsulinUnitsHours).To(Equal("hours"))
	})

	It("BolusCalculatorInsulinUnitsMinutes is expected", func() {
		Expect(pump.BolusCalculatorInsulinUnitsMinutes).To(Equal("minutes"))
	})

	It("BolusCalculatorInsulinUnitsSeconds is expected", func() {
		Expect(pump.BolusCalculatorInsulinUnitsSeconds).To(Equal("seconds"))
	})

	It("BolusCalculatorInsulinUnits returns expected", func() {
		Expect(pump.BolusCalculatorInsulinUnits()).To(Equal([]string{"hours", "minutes", "seconds"}))
	})

	Context("ParseBolusCalculatorInsulin", func() {
		// TODO
	})

	Context("NewBolusCalculatorInsulin", func() {
		It("is successful", func() {
			Expect(pump.NewBolusCalculatorInsulin()).To(Equal(&pump.BolusCalculatorInsulin{}))
		})
	})

	Context("BolusCalculatorInsulin", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.BolusCalculatorInsulin), expectedErrors ...error) {
					datum := pumpTest.NewBolusCalculatorInsulin()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BolusCalculatorInsulin) {},
				),
				Entry("units missing; duration missing",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = nil
						datum.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/duration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; duration out of range (lower)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(-0.1)
						datum.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; duration in range (lower)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(0.0)
						datum.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; duration in range (upper)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(10.0)
						datum.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; duration out of range (upper)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(10.1)
						datum.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; duration missing",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = nil
						datum.Units = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/duration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units invalid; duration out of range (lower)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(-0.1)
						datum.Units = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units invalid; duration in range (lower)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(0.0)
						datum.Units = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units invalid; duration in range (upper)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(10.0)
						datum.Units = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units invalid; duration out of range (upper)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(10.1)
						datum.Units = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units hours: duration missing",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = nil
						datum.Units = pointer.FromString("hours")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/duration"),
				),
				Entry("units hours: duration out of range (lower)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(-0.1)
						datum.Units = pointer.FromString("hours")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10.0), "/duration"),
				),
				Entry("units hours: duration in range (lower)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(0.0)
						datum.Units = pointer.FromString("hours")
					},
				),
				Entry("units hours: duration in range (upper)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(10.0)
						datum.Units = pointer.FromString("hours")
					},
				),
				Entry("units hours: duration out of range (upper)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(10.1)
						datum.Units = pointer.FromString("hours")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10.1, 0.0, 10.0), "/duration"),
				),
				Entry("units minutes: duration missing",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = nil
						datum.Units = pointer.FromString("minutes")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/duration"),
				),
				Entry("units minutes: duration out of range (lower)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(-0.1)
						datum.Units = pointer.FromString("minutes")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 600.0), "/duration"),
				),
				Entry("units minutes: duration in range (lower)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(0.0)
						datum.Units = pointer.FromString("minutes")
					},
				),
				Entry("units minutes: duration in range (upper)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(600.0)
						datum.Units = pointer.FromString("minutes")
					},
				),
				Entry("units minutes: duration out of range (upper)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(600.1)
						datum.Units = pointer.FromString("minutes")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(600.1, 0.0, 600.0), "/duration"),
				),
				Entry("units seconds: duration missing",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = nil
						datum.Units = pointer.FromString("seconds")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/duration"),
				),
				Entry("units seconds: duration out of range (lower)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(-0.1)
						datum.Units = pointer.FromString("seconds")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 36000.0), "/duration"),
				),
				Entry("units seconds: duration in range (lower)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(0.0)
						datum.Units = pointer.FromString("seconds")
					},
				),
				Entry("units seconds: duration in range (upper)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(36000.0)
						datum.Units = pointer.FromString("seconds")
					},
				),
				Entry("units seconds: duration out of range (upper)",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(36000.1)
						datum.Units = pointer.FromString("seconds")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(36000.1, 0.0, 36000.0), "/duration"),
				),
				Entry("units missing",
					func(datum *pump.BolusCalculatorInsulin) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *pump.BolusCalculatorInsulin) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units hours",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = pointer.FromFloat64(0.0)
						datum.Units = pointer.FromString("hours")
					},
				),
				Entry("multiple errors",
					func(datum *pump.BolusCalculatorInsulin) {
						datum.Duration = nil
						datum.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/duration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BolusCalculatorInsulin)) {
					for _, origin := range structure.Origins() {
						datum := pumpTest.NewBolusCalculatorInsulin()
						mutator(datum)
						expectedDatum := pumpTest.CloneBolusCalculatorInsulin(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.BolusCalculatorInsulin) {},
				),
				Entry("does not modify the datum; duration missing",
					func(datum *pump.BolusCalculatorInsulin) { datum.Duration = nil },
				),
				Entry("does not modify the datum; units missing",
					func(datum *pump.BolusCalculatorInsulin) { datum.Units = nil },
				),
			)
		})
	})

	Context("BolusCalculatorInsulinDurationRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := pump.BolusCalculatorInsulinDurationRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := pump.BolusCalculatorInsulinDurationRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units hours", func() {
			minimum, maximum := pump.BolusCalculatorInsulinDurationRangeForUnits(pointer.FromString("hours"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(10.0))
		})

		It("returns expected range for units minutes", func() {
			minimum, maximum := pump.BolusCalculatorInsulinDurationRangeForUnits(pointer.FromString("minutes"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(600.0))
		})

		It("returns expected range for units seconds", func() {
			minimum, maximum := pump.BolusCalculatorInsulinDurationRangeForUnits(pointer.FromString("seconds"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(36000.0))
		})
	})
})
