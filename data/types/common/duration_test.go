package commontypes_test

import (
	"math"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	commontypes "github.com/tidepool-org/platform/data/types/common"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewDuration() *commontypes.Duration {
	datum := commontypes.NewDuration()
	datum.Units = pointer.FromString(test.RandomStringFromArray(commontypes.DurationUnits()))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(commontypes.DurationValueRangeForUnits(datum.Units)))
	return datum
}

func CloneDuration(datum *commontypes.Duration) *commontypes.Duration {
	if datum == nil {
		return nil
	}
	clone := commontypes.NewDuration()
	clone.Units = pointer.CloneString(datum.Units)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}

var _ = Describe("Duration", func() {
	It("DurationUnitsHours is expected", func() {
		Expect(commontypes.DurationUnitsHours).To(Equal("hours"))
	})

	It("DurationUnitsMinutes is expected", func() {
		Expect(commontypes.DurationUnitsMinutes).To(Equal("minutes"))
	})

	It("DurationUnitsSeconds is expected", func() {
		Expect(commontypes.DurationUnitsSeconds).To(Equal("seconds"))
	})

	It("DurationValueHoursMaximum is expected", func() {
		Expect(commontypes.DurationValueHoursMaximum).To(Equal(168.0))
	})

	It("DurationValueHoursMinimum is expected", func() {
		Expect(commontypes.DurationValueHoursMinimum).To(Equal(0.0))
	})

	It("DurationValueMinutesMaximum is expected", func() {
		Expect(commontypes.DurationValueMinutesMaximum).To(Equal(10080.0))
	})

	It("DurationValueMinutesMinimum is expected", func() {
		Expect(commontypes.DurationValueMinutesMinimum).To(Equal(0.0))
	})

	It("DurationValueSecondsMaximum is expected", func() {
		Expect(commontypes.DurationValueSecondsMaximum).To(Equal(604800.0))
	})

	It("DurationValueSecondsMinimum is expected", func() {
		Expect(commontypes.DurationValueSecondsMinimum).To(Equal(0.0))
	})

	It("DurationUnits returns expected", func() {
		Expect(commontypes.DurationUnits()).To(Equal([]string{"hours", "minutes", "seconds"}))
	})

	Context("ParseDuration", func() {
		// TODO
	})

	Context("NewDuration", func() {
		It("returns the expected datum", func() {
			Expect(commontypes.NewDuration()).To(Equal(&commontypes.Duration{}))
		})
	})

	Context("Duration", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *commontypes.Duration), expectedErrors ...error) {
					datum := NewDuration()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *commontypes.Duration) {},
				),
				Entry("units missing; value missing",
					func(datum *commontypes.Duration) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value out of range (lower)",
					func(datum *commontypes.Duration) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *commontypes.Duration) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *commontypes.Duration) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(604800.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value out of range (upper)",
					func(datum *commontypes.Duration) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(604800.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value out of range (lower)",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(604800.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units invalid; value out of range (upper)",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(604800.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units hours; value missing",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("hours")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units hours; value out of range (lower)",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("hours")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 168.0), "/value"),
				),
				Entry("units hours; value in range (lower)",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("hours")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units hours; value in range (upper)",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("hours")
						datum.Value = pointer.FromFloat64(168.0)
					},
				),
				Entry("units hours; value out of range (upper)",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("hours")
						datum.Value = pointer.FromFloat64(168.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(168.1, 0.0, 168.0), "/value"),
				),
				Entry("units minutes; value missing",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("minutes")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units minutes; value out of range (lower)",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("minutes")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10080.0), "/value"),
				),
				Entry("units minutes; value in range (lower)",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("minutes")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units minutes; value in range (upper)",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("minutes")
						datum.Value = pointer.FromFloat64(10080.0)
					},
				),
				Entry("units minutes; value out of range (upper)",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("minutes")
						datum.Value = pointer.FromFloat64(10080.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10080.1, 0.0, 10080.0), "/value"),
				),
				Entry("units seconds; value missing",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("seconds")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units seconds; value out of range (lower)",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("seconds")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 604800.0), "/value"),
				),
				Entry("units seconds; value in range (lower)",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("seconds")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units seconds; value in range (upper)",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("seconds")
						datum.Value = pointer.FromFloat64(604800.0)
					},
				),
				Entry("units seconds; value out of range (upper)",
					func(datum *commontypes.Duration) {
						datum.Units = pointer.FromString("seconds")
						datum.Value = pointer.FromFloat64(604800.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(604800.1, 0.0, 604800.0), "/value"),
				),
				Entry("multiple errors",
					func(datum *commontypes.Duration) {
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
				func(mutator func(datum *commontypes.Duration)) {
					for _, origin := range structure.Origins() {
						datum := NewDuration()
						mutator(datum)
						expectedDatum := CloneDuration(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *commontypes.Duration) {},
				),
				Entry("does not modify the datum; units missing",
					func(datum *commontypes.Duration) { datum.Units = nil },
				),
				Entry("does not modify the datum; units hours",
					func(datum *commontypes.Duration) { datum.Units = pointer.FromString("hours") },
				),
				Entry("does not modify the datum; units minutes",
					func(datum *commontypes.Duration) { datum.Units = pointer.FromString("minutes") },
				),
				Entry("does not modify the datum; units seconds",
					func(datum *commontypes.Duration) { datum.Units = pointer.FromString("seconds") },
				),
				Entry("does not modify the datum; value missing",
					func(datum *commontypes.Duration) { datum.Value = nil },
				),
			)
		})
	})

	Context("DurationValueRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := commontypes.DurationValueRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := commontypes.DurationValueRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units hours", func() {
			minimum, maximum := commontypes.DurationValueRangeForUnits(pointer.FromString("hours"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(168.0))
		})

		It("returns expected range for units minutes", func() {
			minimum, maximum := commontypes.DurationValueRangeForUnits(pointer.FromString("minutes"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(10080.0))
		})

		It("returns expected range for units seconds", func() {
			minimum, maximum := commontypes.DurationValueRangeForUnits(pointer.FromString("seconds"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(604800.0))
		})
	})
})
