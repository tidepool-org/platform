package cgm_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"

	dataTypesCommon "github.com/tidepool-org/platform/data/types/common"
	dataTypesSettingsCgmTest "github.com/tidepool-org/platform/data/types/settings/cgm/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("ScheduledAlert", func() {
	It("ScheduledAlertsLengthMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.ScheduledAlertsLengthMaximum).To(Equal(10))
	})

	It("ScheduledAlertNameLengthMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.ScheduledAlertNameLengthMaximum).To(Equal(100))
	})

	It("ScheduledAlertStartMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.ScheduledAlertStartMaximum).To(Equal(86400000))
	})

	It("ScheduledAlertStartMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.ScheduledAlertStartMinimum).To(Equal(0))
	})

	It("ScheduledAlertEndMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.ScheduledAlertEndMaximum).To(Equal(172800000))
	})

	It("ScheduledAlertEndMinimum is expected", func() {
		Expect(dataTypesSettingsCgm.ScheduledAlertEndMinimum).To(Equal(0))
	})

	Context("ParseScheduledAlerts", func() {
		// TODO
	})

	Context("NewScheduledAlerts", func() {
		It("returns successfully with default values", func() {
			datum := dataTypesSettingsCgm.NewScheduledAlerts()
			Expect(datum).ToNot(BeNil())
			Expect(*datum).To(BeEmpty())
		})
	})

	Context("ScheduledAlerts", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsCgm.ScheduledAlerts), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomScheduledAlerts(1, 3)
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsCgm.ScheduledAlerts) {},
				),
				Entry("empty",
					func(datum *dataTypesSettingsCgm.ScheduledAlerts) { *datum = *dataTypesSettingsCgm.NewScheduledAlerts() },
					structureValidator.ErrorValueEmpty(),
				),
				Entry("length in range (lower)",
					func(datum *dataTypesSettingsCgm.ScheduledAlerts) {
						*datum = *dataTypesSettingsCgmTest.RandomScheduledAlerts(1, 1)
					},
				),
				Entry("length in range (upper)",
					func(datum *dataTypesSettingsCgm.ScheduledAlerts) {
						*datum = *dataTypesSettingsCgmTest.RandomScheduledAlerts(10, 10)
					},
				),
				Entry("length out of range (upper)",
					func(datum *dataTypesSettingsCgm.ScheduledAlerts) {
						*datum = *dataTypesSettingsCgmTest.RandomScheduledAlerts(11, 11)
					},
					structureValidator.ErrorLengthNotLessThanOrEqualTo(11, 10),
				),
				Entry("entry missing",
					func(datum *dataTypesSettingsCgm.ScheduledAlerts) {
						(*datum)[0] = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsCgm.ScheduledAlerts) {
						*datum = *dataTypesSettingsCgmTest.RandomScheduledAlerts(11, 11)
						(*datum)[1] = nil
					},
					structureValidator.ErrorLengthNotLessThanOrEqualTo(11, 10),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1"),
				),
			)
		})
	})

	Context("ParseScheduledAlert", func() {
		// TODO
	})

	Context("NewScheduledAlert", func() {
		It("returns successfully with default values", func() {
			datum := dataTypesSettingsCgm.NewScheduledAlert()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Name).To(BeNil())
			Expect(datum.Days).To(BeNil())
			Expect(datum.Start).To(BeNil())
			Expect(datum.End).To(BeNil())
			Expect(datum.Alerts).To(BeNil())
		})
	})

	Context("ScheduledAlert", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsCgm.ScheduledAlert)) {
				datum := dataTypesSettingsCgmTest.RandomScheduledAlert()
				mutator(datum)
				test.ExpectSerializedObjectBSON(datum, dataTypesSettingsCgmTest.NewObjectFromScheduledAlert(datum, test.ObjectFormatBSON))
				test.ExpectSerializedObjectJSON(datum, dataTypesSettingsCgmTest.NewObjectFromScheduledAlert(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsCgm.ScheduledAlert) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsCgm.ScheduledAlert) { *datum = dataTypesSettingsCgm.ScheduledAlert{} },
			),
		)

		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsCgm.ScheduledAlert), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomScheduledAlert()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) {},
				),
				Entry("name missing",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) { datum.Name = nil },
				),
				Entry("name empty",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) { datum.Name = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
				Entry("name length in range (upper)",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) {
						datum.Name = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("name length out of range (upper)",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) {
						datum.Name = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/name"),
				),
				Entry("days missing",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) { datum.Days = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/days"),
				),
				Entry("days contains invalid",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) {
						datum.Days = pointer.FromStringArray(append([]string{"invalid"}, test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(0, len(dataTypesCommon.DaysOfWeek())-1, dataTypesCommon.DaysOfWeek())...))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}), "/days/0"),
				),
				Entry("days contains duplicate",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) {
						duplicate := test.RandomStringFromArray(dataTypesCommon.DaysOfWeek())
						datum.Days = pointer.FromStringArray([]string{duplicate, duplicate})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/days/1"),
				),
				Entry("days valid",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) {
						datum.Days = pointer.FromStringArray(test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(1, len(dataTypesCommon.DaysOfWeek()), dataTypesCommon.DaysOfWeek()))
					},
				),
				Entry("start missing",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) { datum.Start = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
				Entry("start out of range (lower)",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) { datum.Start = pointer.FromInt(-1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/start"),
				),
				Entry("start in range (lower)",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) { datum.Start = pointer.FromInt(0) },
				),
				Entry("start in range (upper)",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) { datum.Start = pointer.FromInt(86400000) },
				),
				Entry("start out of range (upper)",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) { datum.Start = pointer.FromInt(86400001) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/start"),
				),
				Entry("end missing",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) { datum.End = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/end"),
				),
				Entry("end out of range (lower)",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) { datum.End = pointer.FromInt(-1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 172800000), "/end"),
				),
				Entry("end in range (lower)",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) { datum.End = pointer.FromInt(0) },
				),
				Entry("end in range (upper)",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) { datum.End = pointer.FromInt(172800000) },
				),
				Entry("end out of range (upper)",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) { datum.End = pointer.FromInt(172800001) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(172800001, 0, 172800000), "/end"),
				),
				Entry("alerts missing",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) { datum.Alerts = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alerts"),
				),
				Entry("alerts invalid",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) { datum.Alerts.UrgentLow.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alerts/urgentLow/enabled"),
				),
				Entry("alerts valid",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) {
						datum.Alerts = dataTypesSettingsCgmTest.RandomAlerts()
					},
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsCgm.ScheduledAlert) {
						datum.Name = pointer.FromString("")
						datum.Days = nil
						datum.Start = nil
						datum.End = nil
						datum.Alerts = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/days"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/end"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alerts"),
				),
			)
		})
	})
})
