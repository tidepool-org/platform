package cgm_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	dataTypesSettingsCgmTest "github.com/tidepool-org/platform/data/types/settings/cgm/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Snooze", func() {
	It("SnoozeUnitsHours is expected", func() {
		Expect(dataTypesSettingsCgm.SnoozeUnitsHours).To(Equal("hours"))
	})

	It("SnoozeUnitsMinutes is expected", func() {
		Expect(dataTypesSettingsCgm.SnoozeUnitsMinutes).To(Equal("minutes"))
	})

	It("SnoozeUnitsSeconds is expected", func() {
		Expect(dataTypesSettingsCgm.SnoozeUnitsSeconds).To(Equal("seconds"))
	})

	It("SnoozeDurationHoursMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.SnoozeDurationHoursMaximum).To(Equal(10.0))
	})

	It("SnoozeDurationHoursMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.SnoozeDurationHoursMinimum).To(Equal(0.0))
	})

	It("SnoozeDurationMinutesMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.SnoozeDurationMinutesMaximum).To(Equal(600.0))
	})

	It("SnoozeDurationMinutesMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.SnoozeDurationMinutesMinimum).To(Equal(0.0))
	})

	It("SnoozeDurationSecondsMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.SnoozeDurationSecondsMaximum).To(Equal(36000.0))
	})

	It("SnoozeDurationSecondsMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.SnoozeDurationSecondsMinimum).To(Equal(0.0))
	})

	It("SnoozeUnits returns expected", func() {
		Expect(dataTypesSettingsCgm.SnoozeUnits()).To(Equal([]string{"hours", "minutes", "seconds"}))
	})

	Context("ParseSnooze", func() {
		// TODO
	})

	Context("NewSnooze", func() {
		It("returns successfully with default values", func() {
			datum := dataTypesSettingsCgm.NewSnooze()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Duration).To(BeNil())
			Expect(datum.Units).To(BeNil())
		})
	})

	Context("Snooze", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsCgm.Snooze)) {
				datum := dataTypesSettingsCgmTest.RandomSnooze()
				mutator(datum)
				test.ExpectSerializedObjectBSON(datum, dataTypesSettingsCgmTest.NewObjectFromSnooze(datum, test.ObjectFormatBSON))
				test.ExpectSerializedObjectJSON(datum, dataTypesSettingsCgmTest.NewObjectFromSnooze(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsCgm.Snooze) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsCgm.Snooze) { *datum = dataTypesSettingsCgm.Snooze{} },
			),
		)

		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsCgm.Snooze), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomSnooze()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsCgm.Snooze) {},
				),
				Entry("units missing; duration missing",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = nil
						datum.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/duration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; duration exists",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = nil
						datum.Duration = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; duration missing",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = pointer.FromString("invalid")
						datum.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/duration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units invalid; duration exists",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = pointer.FromString("invalid")
						datum.Duration = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units hours; duration missing",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = pointer.FromString("hours")
						datum.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/duration"),
				),
				Entry("units hours; duration out of range (lower)",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = pointer.FromString("hours")
						datum.Duration = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10.0), "/duration"),
				),
				Entry("units hours; duration in range (lower)",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = pointer.FromString("hours")
						datum.Duration = pointer.FromFloat64(0.0)
					},
				),
				Entry("units hours; duration in range (upper)",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = pointer.FromString("hours")
						datum.Duration = pointer.FromFloat64(10.0)
					},
				),
				Entry("units hours; duration out of range (upper)",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = pointer.FromString("hours")
						datum.Duration = pointer.FromFloat64(10.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10.1, 0.0, 10.0), "/duration"),
				),
				Entry("units minutes; duration missing",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = pointer.FromString("minutes")
						datum.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/duration"),
				),
				Entry("units minutes; duration out of range (lower)",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = pointer.FromString("minutes")
						datum.Duration = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 600.0), "/duration"),
				),
				Entry("units minutes; duration in range (lower)",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = pointer.FromString("minutes")
						datum.Duration = pointer.FromFloat64(0.0)
					},
				),
				Entry("units minutes; duration in range (upper)",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = pointer.FromString("minutes")
						datum.Duration = pointer.FromFloat64(600.0)
					},
				),
				Entry("units minutes; duration out of range (upper)",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = pointer.FromString("minutes")
						datum.Duration = pointer.FromFloat64(600.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(600.1, 0.0, 600.0), "/duration"),
				),
				Entry("units seconds; duration missing",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = pointer.FromString("seconds")
						datum.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/duration"),
				),
				Entry("units seconds; duration out of range (lower)",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = pointer.FromString("seconds")
						datum.Duration = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 36000.0), "/duration"),
				),
				Entry("units seconds; duration in range (lower)",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = pointer.FromString("seconds")
						datum.Duration = pointer.FromFloat64(0.0)
					},
				),
				Entry("units seconds; duration in range (upper)",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = pointer.FromString("seconds")
						datum.Duration = pointer.FromFloat64(36000.0)
					},
				),
				Entry("units seconds; duration out of range (upper)",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = pointer.FromString("seconds")
						datum.Duration = pointer.FromFloat64(36000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(36000.1, 0.0, 36000.0), "/duration"),
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsCgm.Snooze) {
						datum.Units = nil
						datum.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/duration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})
	})

	Context("SnoozeDurationRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := dataTypesSettingsCgm.SnoozeDurationRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := dataTypesSettingsCgm.SnoozeDurationRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units hours", func() {
			minimum, maximum := dataTypesSettingsCgm.SnoozeDurationRangeForUnits(pointer.FromString("hours"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(10.0))
		})

		It("returns expected range for units minutes", func() {
			minimum, maximum := dataTypesSettingsCgm.SnoozeDurationRangeForUnits(pointer.FromString("minutes"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(600.0))
		})

		It("returns expected range for units seconds", func() {
			minimum, maximum := dataTypesSettingsCgm.SnoozeDurationRangeForUnits(pointer.FromString("seconds"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(36000.0))
		})
	})
})
