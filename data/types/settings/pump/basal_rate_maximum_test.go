package pump_test

import (
	"math"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	pumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("BasalRateMaximum", func() {
	It("BasalRateMaximumUnitsUnitsPerHour is expected", func() {
		Expect(pump.BasalRateMaximumUnitsUnitsPerHour).To(Equal("Units/hour"))
	})

	It("BasalRateMaximumValueUnitsPerHourMaximum is expected", func() {
		Expect(pump.BasalRateMaximumValueUnitsPerHourMaximum).To(Equal(100.0))
	})

	It("BasalRateMaximumValueUnitsPerHourMinimum is expected", func() {
		Expect(pump.BasalRateMaximumValueUnitsPerHourMinimum).To(Equal(0.0))
	})

	It("BasalRateMaximumUnits returns expected", func() {
		Expect(pump.BasalRateMaximumUnits()).To(Equal([]string{"Units/hour"}))
	})

	Context("ParseBasalRateMaximum", func() {
		// TODO
	})

	Context("NewBasalRateMaximum", func() {
		It("is successful", func() {
			Expect(pump.NewBasalRateMaximum()).To(Equal(&pump.BasalRateMaximum{}))
		})
	})

	Context("BasalRateMaximum", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.BasalRateMaximum), expectedErrors ...error) {
					datum := pumpTest.NewBasalRateMaximum()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BasalRateMaximum) {},
				),
				Entry("units missing",
					func(datum *pump.BasalRateMaximum) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *pump.BasalRateMaximum) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units/hour"}), "/units"),
				),
				Entry("units Units/hour",
					func(datum *pump.BasalRateMaximum) {
						datum.Units = pointer.FromString("Units/hour")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units missing; value missing",
					func(datum *pump.BasalRateMaximum) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value out of range (lower)",
					func(datum *pump.BasalRateMaximum) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *pump.BasalRateMaximum) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *pump.BasalRateMaximum) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value out of range (upper)",
					func(datum *pump.BasalRateMaximum) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *pump.BasalRateMaximum) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units/hour"}), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value out of range (lower)",
					func(datum *pump.BasalRateMaximum) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units/hour"}), "/units"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *pump.BasalRateMaximum) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units/hour"}), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *pump.BasalRateMaximum) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units/hour"}), "/units"),
				),
				Entry("units invalid; value out of range (upper)",
					func(datum *pump.BasalRateMaximum) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units/hour"}), "/units"),
				),
				Entry("units Units/hour; value missing",
					func(datum *pump.BasalRateMaximum) {
						datum.Units = pointer.FromString("Units/hour")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units Units/hour; value out of range (lower)",
					func(datum *pump.BasalRateMaximum) {
						datum.Units = pointer.FromString("Units/hour")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/value"),
				),
				Entry("units Units/hour; value in range (lower)",
					func(datum *pump.BasalRateMaximum) {
						datum.Units = pointer.FromString("Units/hour")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units Units/hour; value in range (upper)",
					func(datum *pump.BasalRateMaximum) {
						datum.Units = pointer.FromString("Units/hour")
						datum.Value = pointer.FromFloat64(100.0)
					},
				),
				Entry("units Units/hour; value out of range (upper)",
					func(datum *pump.BasalRateMaximum) {
						datum.Units = pointer.FromString("Units/hour")
						datum.Value = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/value"),
				),
				Entry("multiple errors",
					func(datum *pump.BasalRateMaximum) {
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
				func(mutator func(datum *pump.BasalRateMaximum)) {
					for _, origin := range structure.Origins() {
						datum := pumpTest.NewBasalRateMaximum()
						mutator(datum)
						expectedDatum := pumpTest.CloneBasalRateMaximum(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.BasalRateMaximum) {},
				),
				Entry("does not modify the datum; units missing",
					func(datum *pump.BasalRateMaximum) { datum.Units = nil },
				),
				Entry("does not modify the datum; value missing",
					func(datum *pump.BasalRateMaximum) { datum.Value = nil },
				),
			)
		})
	})

	Context("BasalRateMaximumValueRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := pump.BasalRateMaximumValueRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := pump.BasalRateMaximumValueRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units Units/hour", func() {
			minimum, maximum := pump.BasalRateMaximumValueRangeForUnits(pointer.FromString("Units/hour"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(100.0))
		})
	})
})
