package common_test

import (
	"math"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesCommon "github.com/tidepool-org/platform/data/types/common"
	dataTypesCommonTest "github.com/tidepool-org/platform/data/types/common/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Duration", func() {
	It("DurationUnitsHours is expected", func() {
		Expect(dataTypesCommon.DurationUnitsHours).To(Equal("hours"))
	})

	It("DurationUnitsMinutes is expected", func() {
		Expect(dataTypesCommon.DurationUnitsMinutes).To(Equal("minutes"))
	})

	It("DurationUnitsSeconds is expected", func() {
		Expect(dataTypesCommon.DurationUnitsSeconds).To(Equal("seconds"))
	})

	It("DurationValueHoursMaximum is expected", func() {
		Expect(dataTypesCommon.DurationValueHoursMaximum).To(Equal(168.0))
	})

	It("DurationValueHoursMinimum is expected", func() {
		Expect(dataTypesCommon.DurationValueHoursMinimum).To(Equal(0.0))
	})

	It("DurationValueMinutesMaximum is expected", func() {
		Expect(dataTypesCommon.DurationValueMinutesMaximum).To(Equal(10080.0))
	})

	It("DurationValueMinutesMinimum is expected", func() {
		Expect(dataTypesCommon.DurationValueMinutesMinimum).To(Equal(0.0))
	})

	It("DurationValueSecondsMaximum is expected", func() {
		Expect(dataTypesCommon.DurationValueSecondsMaximum).To(Equal(604800.0))
	})

	It("DurationValueSecondsMinimum is expected", func() {
		Expect(dataTypesCommon.DurationValueSecondsMinimum).To(Equal(0.0))
	})

	It("DurationUnits returns expected", func() {
		Expect(dataTypesCommon.DurationUnits()).To(Equal([]string{"hours", "minutes", "seconds"}))
	})

	Context("ParseDuration", func() {
		// TODO
	})

	Context("NewDuration", func() {
		It("returns the expected datum", func() {
			Expect(dataTypesCommon.NewDuration()).To(Equal(&dataTypesCommon.Duration{}))
		})
	})

	Context("Duration", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesCommon.Duration), expectedErrors ...error) {
					datum := dataTypesCommonTest.NewDuration()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesCommon.Duration) {},
				),
				Entry("units missing; value missing",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value out of range (lower)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(604800.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value out of range (upper)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(604800.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value out of range (lower)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(604800.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units invalid; value out of range (upper)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(604800.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units hours; value missing",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("hours")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units hours; value out of range (lower)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("hours")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 168.0), "/value"),
				),
				Entry("units hours; value in range (lower)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("hours")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units hours; value in range (upper)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("hours")
						datum.Value = pointer.FromFloat64(168.0)
					},
				),
				Entry("units hours; value out of range (upper)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("hours")
						datum.Value = pointer.FromFloat64(168.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(168.1, 0.0, 168.0), "/value"),
				),
				Entry("units minutes; value missing",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("minutes")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units minutes; value out of range (lower)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("minutes")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10080.0), "/value"),
				),
				Entry("units minutes; value in range (lower)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("minutes")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units minutes; value in range (upper)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("minutes")
						datum.Value = pointer.FromFloat64(10080.0)
					},
				),
				Entry("units minutes; value out of range (upper)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("minutes")
						datum.Value = pointer.FromFloat64(10080.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10080.1, 0.0, 10080.0), "/value"),
				),
				Entry("units seconds; value missing",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("seconds")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units seconds; value out of range (lower)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("seconds")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 604800.0), "/value"),
				),
				Entry("units seconds; value in range (lower)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("seconds")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units seconds; value in range (upper)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("seconds")
						datum.Value = pointer.FromFloat64(604800.0)
					},
				),
				Entry("units seconds; value out of range (upper)",
					func(datum *dataTypesCommon.Duration) {
						datum.Units = pointer.FromString("seconds")
						datum.Value = pointer.FromFloat64(604800.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(604800.1, 0.0, 604800.0), "/value"),
				),
				Entry("multiple errors",
					func(datum *dataTypesCommon.Duration) {
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
				func(mutator func(datum *dataTypesCommon.Duration)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesCommonTest.NewDuration()
						mutator(datum)
						expectedDatum := dataTypesCommonTest.CloneDuration(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *dataTypesCommon.Duration) {},
				),
				Entry("does not modify the datum; units missing",
					func(datum *dataTypesCommon.Duration) { datum.Units = nil },
				),
				Entry("does not modify the datum; units hours",
					func(datum *dataTypesCommon.Duration) { datum.Units = pointer.FromString("hours") },
				),
				Entry("does not modify the datum; units minutes",
					func(datum *dataTypesCommon.Duration) { datum.Units = pointer.FromString("minutes") },
				),
				Entry("does not modify the datum; units seconds",
					func(datum *dataTypesCommon.Duration) { datum.Units = pointer.FromString("seconds") },
				),
				Entry("does not modify the datum; value missing",
					func(datum *dataTypesCommon.Duration) { datum.Value = nil },
				),
			)
		})
	})

	Context("DurationValueRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := dataTypesCommon.DurationValueRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := dataTypesCommon.DurationValueRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units hours", func() {
			minimum, maximum := dataTypesCommon.DurationValueRangeForUnits(pointer.FromString("hours"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(168.0))
		})

		It("returns expected range for units minutes", func() {
			minimum, maximum := dataTypesCommon.DurationValueRangeForUnits(pointer.FromString("minutes"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(10080.0))
		})

		It("returns expected range for units seconds", func() {
			minimum, maximum := dataTypesCommon.DurationValueRangeForUnits(pointer.FromString("seconds"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(604800.0))
		})
	})
})
