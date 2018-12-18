package cgm_test

import (
	"math"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataTest "github.com/tidepool-org/platform/data/test"
	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	dataTypesSettingsCgmTest "github.com/tidepool-org/platform/data/types/settings/cgm/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("DurationAlert", func() {
	It("DurationAlertUnitsHours is expected", func() {
		Expect(dataTypesSettingsCgm.DurationAlertUnitsHours).To(Equal("hours"))
	})

	It("DurationAlertUnitsMinutes is expected", func() {
		Expect(dataTypesSettingsCgm.DurationAlertUnitsMinutes).To(Equal("minutes"))
	})

	It("DurationAlertUnitsSeconds is expected", func() {
		Expect(dataTypesSettingsCgm.DurationAlertUnitsSeconds).To(Equal("seconds"))
	})

	It("NoDataAlertDurationHoursMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.NoDataAlertDurationHoursMaximum).To(Equal(6.0))
	})

	It("NoDataAlertDurationHoursMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.NoDataAlertDurationHoursMinimum).To(Equal(0.0))
	})

	It("NoDataAlertDurationMinutesMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.NoDataAlertDurationMinutesMaximum).To(Equal(360.0))
	})

	It("NoDataAlertDurationMinutesMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.NoDataAlertDurationMinutesMinimum).To(Equal(0.0))
	})
	It("NoDataAlertDurationSecondsMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.NoDataAlertDurationSecondsMaximum).To(Equal(21600.0))
	})

	It("NoDataAlertDurationSecondsMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.NoDataAlertDurationSecondsMinimum).To(Equal(0.0))
	})

	It("OutOfRangeAlertDurationHoursMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.OutOfRangeAlertDurationHoursMaximum).To(Equal(6.0))
	})

	It("OutOfRangeAlertDurationHoursMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.OutOfRangeAlertDurationHoursMinimum).To(Equal(0.0))
	})

	It("OutOfRangeAlertDurationMinutesMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.OutOfRangeAlertDurationMinutesMaximum).To(Equal(360.0))
	})

	It("OutOfRangeAlertDurationMinutesMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.OutOfRangeAlertDurationMinutesMinimum).To(Equal(0.0))
	})
	It("OutOfRangeAlertDurationSecondsMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.OutOfRangeAlertDurationSecondsMaximum).To(Equal(21600.0))
	})

	It("OutOfRangeAlertDurationSecondsMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.OutOfRangeAlertDurationSecondsMinimum).To(Equal(0.0))
	})

	It("DurationAlertUnits returns expected", func() {
		Expect(dataTypesSettingsCgm.DurationAlertUnits()).To(Equal([]string{"hours", "minutes", "seconds"}))
	})

	Context("DurationAlert", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsCgm.DurationAlert)) {
				datum := dataTypesSettingsCgmTest.RandomDurationAlert()
				mutator(datum)
				test.ExpectSerializedObjectBSON(datum, dataTypesSettingsCgmTest.NewObjectFromDurationAlert(datum, test.ObjectFormatBSON))
				test.ExpectSerializedObjectJSON(datum, dataTypesSettingsCgmTest.NewObjectFromDurationAlert(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsCgm.DurationAlert) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsCgm.DurationAlert) { *datum = dataTypesSettingsCgm.DurationAlert{} },
			),
		)

		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsCgm.DurationAlert), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomDurationAlert()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsCgm.DurationAlert) {},
				),
				Entry("enabled missing",
					func(datum *dataTypesSettingsCgm.DurationAlert) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					func(datum *dataTypesSettingsCgm.DurationAlert) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("enabled true",
					func(datum *dataTypesSettingsCgm.DurationAlert) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("snooze missing",
					func(datum *dataTypesSettingsCgm.DurationAlert) { datum.Snooze = nil },
				),
				Entry("snooze invalid",
					func(datum *dataTypesSettingsCgm.DurationAlert) { datum.Snooze.Duration = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
				),
				Entry("snooze valid",
					func(datum *dataTypesSettingsCgm.DurationAlert) {
						datum.Snooze = dataTypesSettingsCgmTest.RandomSnooze()
					},
				),
				Entry("units missing; duration missing",
					func(datum *dataTypesSettingsCgm.DurationAlert) {
						datum.Units = nil
						datum.Duration = nil
					},
				),
				Entry("units missing; duration exists",
					func(datum *dataTypesSettingsCgm.DurationAlert) {
						datum.Units = nil
						datum.Duration = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; duration missing",
					func(datum *dataTypesSettingsCgm.DurationAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units invalid; duration exists",
					func(datum *dataTypesSettingsCgm.DurationAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Duration = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units hours; duration missing",
					func(datum *dataTypesSettingsCgm.DurationAlert) {
						datum.Units = pointer.FromString("hours")
						datum.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units hours; duration exists",
					func(datum *dataTypesSettingsCgm.DurationAlert) {
						datum.Units = pointer.FromString("hours")
						datum.Duration = pointer.FromFloat64(test.RandomFloat64())
					},
				),
				Entry("units minutes; duration missing",
					func(datum *dataTypesSettingsCgm.DurationAlert) {
						datum.Units = pointer.FromString("minutes")
						datum.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units minutes; duration exists",
					func(datum *dataTypesSettingsCgm.DurationAlert) {
						datum.Units = pointer.FromString("minutes")
						datum.Duration = pointer.FromFloat64(test.RandomFloat64())
					},
				),
				Entry("units seconds; duration missing",
					func(datum *dataTypesSettingsCgm.DurationAlert) {
						datum.Units = pointer.FromString("seconds")
						datum.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units seconds; duration exists",
					func(datum *dataTypesSettingsCgm.DurationAlert) {
						datum.Units = pointer.FromString("seconds")
						datum.Duration = pointer.FromFloat64(test.RandomFloat64())
					},
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsCgm.DurationAlert) {
						datum.Enabled = nil
						datum.Snooze.Duration = nil
						datum.Units = nil
						datum.Duration = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})
	})

	Context("ParseNoDataAlert", func() {
		// TODO
	})

	Context("NewNoDataAlert", func() {
		It("returns successfully with default values", func() {
			datum := dataTypesSettingsCgm.NewNoDataAlert()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Enabled).To(BeNil())
			Expect(datum.Snooze).To(BeNil())
			Expect(datum.Duration).To(BeNil())
			Expect(datum.Units).To(BeNil())
		})
	})

	Context("NoDataAlert", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsCgm.NoDataAlert)) {
				datum := dataTypesSettingsCgmTest.RandomNoDataAlert()
				mutator(datum)
				test.ExpectSerializedObjectBSON(datum, dataTypesSettingsCgmTest.NewObjectFromNoDataAlert(datum, test.ObjectFormatBSON))
				test.ExpectSerializedObjectJSON(datum, dataTypesSettingsCgmTest.NewObjectFromNoDataAlert(datum, test.ObjectFormatJSON))
				dataTest.ExpectSerializedObject(datum, dataTypesSettingsCgmTest.NewObjectFromNoDataAlert(datum, test.ObjectFormatJSON),
					func(parser data.ObjectParser) interface{} { return dataTypesSettingsCgm.ParseNoDataAlert(parser) })
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsCgm.NoDataAlert) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsCgm.NoDataAlert) { *datum = dataTypesSettingsCgm.NoDataAlert{} },
			),
		)

		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsCgm.NoDataAlert), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomNoDataAlert()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {},
				),
				Entry("enabled missing",
					func(datum *dataTypesSettingsCgm.NoDataAlert) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					func(datum *dataTypesSettingsCgm.NoDataAlert) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("enabled true",
					func(datum *dataTypesSettingsCgm.NoDataAlert) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("snooze missing",
					func(datum *dataTypesSettingsCgm.NoDataAlert) { datum.Snooze = nil },
				),
				Entry("snooze invalid",
					func(datum *dataTypesSettingsCgm.NoDataAlert) { datum.Snooze.Duration = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
				),
				Entry("snooze valid",
					func(datum *dataTypesSettingsCgm.NoDataAlert) { datum.Snooze = dataTypesSettingsCgmTest.RandomSnooze() },
				),
				Entry("units missing; duration missing",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = nil
						datum.Duration = nil
					},
				),
				Entry("units missing; duration exists",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = nil
						datum.Duration = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; duration missing",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units invalid; duration exists",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Duration = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units hours; duration missing",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = pointer.FromString("hours")
						datum.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units hours; duration out of range (lower)",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = pointer.FromString("hours")
						datum.Duration = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 6.0), "/duration"),
				),
				Entry("units hours; duration in range (lower)",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = pointer.FromString("hours")
						datum.Duration = pointer.FromFloat64(0.0)
					},
				),
				Entry("units hours; duration in range (upper)",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = pointer.FromString("hours")
						datum.Duration = pointer.FromFloat64(6.0)
					},
				),
				Entry("units hours; duration out of range (upper)",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = pointer.FromString("hours")
						datum.Duration = pointer.FromFloat64(6.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(6.1, 0.0, 6.0), "/duration"),
				),
				Entry("units minutes; duration missing",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = pointer.FromString("minutes")
						datum.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units minutes; duration out of range (lower)",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = pointer.FromString("minutes")
						datum.Duration = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 360.0), "/duration"),
				),
				Entry("units minutes; duration in range (lower)",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = pointer.FromString("minutes")
						datum.Duration = pointer.FromFloat64(0.0)
					},
				),
				Entry("units minutes; duration in range (upper)",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = pointer.FromString("minutes")
						datum.Duration = pointer.FromFloat64(360.0)
					},
				),
				Entry("units minutes; duration out of range (upper)",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = pointer.FromString("minutes")
						datum.Duration = pointer.FromFloat64(360.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(360.1, 0.0, 360.0), "/duration"),
				),
				Entry("units seconds; duration missing",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = pointer.FromString("seconds")
						datum.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units seconds; duration out of range (lower)",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = pointer.FromString("seconds")
						datum.Duration = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 21600.0), "/duration"),
				),
				Entry("units seconds; duration in range (lower)",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = pointer.FromString("seconds")
						datum.Duration = pointer.FromFloat64(0.0)
					},
				),
				Entry("units seconds; duration in range (upper)",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = pointer.FromString("seconds")
						datum.Duration = pointer.FromFloat64(21600.0)
					},
				),
				Entry("units seconds; duration out of range (upper)",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Units = pointer.FromString("seconds")
						datum.Duration = pointer.FromFloat64(21600.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(21600.1, 0.0, 21600.0), "/duration"),
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsCgm.NoDataAlert) {
						datum.Enabled = nil
						datum.Snooze.Duration = nil
						datum.Units = nil
						datum.Duration = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})
	})

	Context("NoDataAlertDurationRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := dataTypesSettingsCgm.NoDataAlertDurationRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := dataTypesSettingsCgm.NoDataAlertDurationRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units hours", func() {
			minimum, maximum := dataTypesSettingsCgm.NoDataAlertDurationRangeForUnits(pointer.FromString("hours"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(6.0))
		})

		It("returns expected range for units minutes", func() {
			minimum, maximum := dataTypesSettingsCgm.NoDataAlertDurationRangeForUnits(pointer.FromString("minutes"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(360.0))
		})

		It("returns expected range for units seconds", func() {
			minimum, maximum := dataTypesSettingsCgm.NoDataAlertDurationRangeForUnits(pointer.FromString("seconds"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(21600.0))
		})
	})

	Context("ParseOutOfRangeAlert", func() {
		// TODO
	})

	Context("NewOutOfRangeAlert", func() {
		It("returns successfully with default values", func() {
			datum := dataTypesSettingsCgm.NewOutOfRangeAlert()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Enabled).To(BeNil())
			Expect(datum.Snooze).To(BeNil())
			Expect(datum.Duration).To(BeNil())
			Expect(datum.Units).To(BeNil())
		})
	})

	Context("OutOfRangeAlert", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsCgm.OutOfRangeAlert)) {
				datum := dataTypesSettingsCgmTest.RandomOutOfRangeAlert()
				mutator(datum)
				test.ExpectSerializedObjectBSON(datum, dataTypesSettingsCgmTest.NewObjectFromOutOfRangeAlert(datum, test.ObjectFormatBSON))
				test.ExpectSerializedObjectJSON(datum, dataTypesSettingsCgmTest.NewObjectFromOutOfRangeAlert(datum, test.ObjectFormatJSON))
				dataTest.ExpectSerializedObject(datum, dataTypesSettingsCgmTest.NewObjectFromOutOfRangeAlert(datum, test.ObjectFormatJSON),
					func(parser data.ObjectParser) interface{} { return dataTypesSettingsCgm.ParseOutOfRangeAlert(parser) })
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsCgm.OutOfRangeAlert) { *datum = dataTypesSettingsCgm.OutOfRangeAlert{} },
			),
		)

		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsCgm.OutOfRangeAlert), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomOutOfRangeAlert()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {},
				),
				Entry("enabled missing",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("enabled true",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("snooze missing",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) { datum.Snooze = nil },
				),
				Entry("snooze invalid",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) { datum.Snooze.Duration = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
				),
				Entry("snooze valid",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Snooze = dataTypesSettingsCgmTest.RandomSnooze()
					},
				),
				Entry("units missing; duration missing",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = nil
						datum.Duration = nil
					},
				),
				Entry("units missing; duration exists",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = nil
						datum.Duration = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; duration missing",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units invalid; duration exists",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = pointer.FromString("invalid")
						datum.Duration = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"hours", "minutes", "seconds"}), "/units"),
				),
				Entry("units hours; duration missing",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = pointer.FromString("hours")
						datum.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units hours; duration out of range (lower)",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = pointer.FromString("hours")
						datum.Duration = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 6.0), "/duration"),
				),
				Entry("units hours; duration in range (lower)",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = pointer.FromString("hours")
						datum.Duration = pointer.FromFloat64(0.0)
					},
				),
				Entry("units hours; duration in range (upper)",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = pointer.FromString("hours")
						datum.Duration = pointer.FromFloat64(6.0)
					},
				),
				Entry("units hours; duration out of range (upper)",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = pointer.FromString("hours")
						datum.Duration = pointer.FromFloat64(6.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(6.1, 0.0, 6.0), "/duration"),
				),
				Entry("units minutes; duration missing",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = pointer.FromString("minutes")
						datum.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units minutes; duration out of range (lower)",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = pointer.FromString("minutes")
						datum.Duration = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 360.0), "/duration"),
				),
				Entry("units minutes; duration in range (lower)",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = pointer.FromString("minutes")
						datum.Duration = pointer.FromFloat64(0.0)
					},
				),
				Entry("units minutes; duration in range (upper)",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = pointer.FromString("minutes")
						datum.Duration = pointer.FromFloat64(360.0)
					},
				),
				Entry("units minutes; duration out of range (upper)",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = pointer.FromString("minutes")
						datum.Duration = pointer.FromFloat64(360.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(360.1, 0.0, 360.0), "/duration"),
				),
				Entry("units seconds; duration missing",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = pointer.FromString("seconds")
						datum.Duration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("units seconds; duration out of range (lower)",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = pointer.FromString("seconds")
						datum.Duration = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 21600.0), "/duration"),
				),
				Entry("units seconds; duration in range (lower)",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = pointer.FromString("seconds")
						datum.Duration = pointer.FromFloat64(0.0)
					},
				),
				Entry("units seconds; duration in range (upper)",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = pointer.FromString("seconds")
						datum.Duration = pointer.FromFloat64(21600.0)
					},
				),
				Entry("units seconds; duration out of range (upper)",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Units = pointer.FromString("seconds")
						datum.Duration = pointer.FromFloat64(21600.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(21600.1, 0.0, 21600.0), "/duration"),
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsCgm.OutOfRangeAlert) {
						datum.Enabled = nil
						datum.Snooze.Duration = nil
						datum.Units = nil
						datum.Duration = pointer.FromFloat64(test.RandomFloat64())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/snooze/duration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})
	})

	Context("OutOfRangeAlertDurationRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := dataTypesSettingsCgm.OutOfRangeAlertDurationRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := dataTypesSettingsCgm.OutOfRangeAlertDurationRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units hours", func() {
			minimum, maximum := dataTypesSettingsCgm.OutOfRangeAlertDurationRangeForUnits(pointer.FromString("hours"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(6.0))
		})

		It("returns expected range for units minutes", func() {
			minimum, maximum := dataTypesSettingsCgm.OutOfRangeAlertDurationRangeForUnits(pointer.FromString("minutes"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(360.0))
		})

		It("returns expected range for units seconds", func() {
			minimum, maximum := dataTypesSettingsCgm.OutOfRangeAlertDurationRangeForUnits(pointer.FromString("seconds"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(21600.0))
		})
	})
})
