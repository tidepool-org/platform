package dexcom_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Alert", func() {
	It("AlertScheduleSettingsStartTimeDefault is expected", func() {
		Expect(dexcom.AlertScheduleSettingsStartTimeDefault).To(Equal("00:00"))
	})

	It("AlertScheduleSettingsEndTimeDefault is expected", func() {
		Expect(dexcom.AlertScheduleSettingsEndTimeDefault).To(Equal("00:00"))
	})

	It("AlertScheduleSettingsDaySunday is expected", func() {
		Expect(dexcom.AlertScheduleSettingsDaySunday).To(Equal("sunday"))
	})

	It("AlertScheduleSettingsDayMonday is expected", func() {
		Expect(dexcom.AlertScheduleSettingsDayMonday).To(Equal("monday"))
	})

	It("AlertScheduleSettingsDayTuesday is expected", func() {
		Expect(dexcom.AlertScheduleSettingsDayTuesday).To(Equal("tuesday"))
	})

	It("AlertScheduleSettingsDayWednesday is expected", func() {
		Expect(dexcom.AlertScheduleSettingsDayWednesday).To(Equal("wednesday"))
	})

	It("AlertScheduleSettingsDayThursday is expected", func() {
		Expect(dexcom.AlertScheduleSettingsDayThursday).To(Equal("thursday"))
	})

	It("AlertScheduleSettingsDayFriday is expected", func() {
		Expect(dexcom.AlertScheduleSettingsDayFriday).To(Equal("friday"))
	})

	It("AlertScheduleSettingsDaySaturday is expected", func() {
		Expect(dexcom.AlertScheduleSettingsDaySaturday).To(Equal("saturday"))
	})

	It("AlertSettingAlertNameFall is expected", func() {
		Expect(dexcom.AlertSettingAlertNameFall).To(Equal("fall"))
	})

	It("AlertSettingAlertNameHigh is expected", func() {
		Expect(dexcom.AlertSettingAlertNameHigh).To(Equal("high"))
	})

	It("AlertSettingAlertNameLow is expected", func() {
		Expect(dexcom.AlertSettingAlertNameLow).To(Equal("low"))
	})

	It("AlertSettingAlertNameNoReadings is expected", func() {
		Expect(dexcom.AlertSettingAlertNameNoReadings).To(Equal("noReadings"))
	})

	It("AlertSettingAlertNameOutOfRange is expected", func() {
		Expect(dexcom.AlertSettingAlertNameOutOfRange).To(Equal("outOfRange"))
	})

	It("AlertSettingAlertNameRise is expected", func() {
		Expect(dexcom.AlertSettingAlertNameRise).To(Equal("rise"))
	})

	It("AlertSettingAlertNameUrgentLow is expected", func() {
		Expect(dexcom.AlertSettingAlertNameUrgentLow).To(Equal("urgentLow"))
	})

	It("AlertSettingAlertNameUrgentLowSoon is expected", func() {
		Expect(dexcom.AlertSettingAlertNameUrgentLowSoon).To(Equal("urgentLowSoon"))
	})

	It("AlertSettingSnoozeMinutesMaximum is expected", func() {
		Expect(dexcom.AlertSettingSnoozeMinutesMaximum).To(Equal(600.0))
	})

	It("AlertSettingSnoozeMinutesMinimum is expected", func() {
		Expect(dexcom.AlertSettingSnoozeMinutesMinimum).To(Equal(0.0))
	})

	It("AlertSettingUnitMinutes is expected", func() {
		Expect(dexcom.AlertSettingUnitMinutes).To(Equal("minutes"))
	})

	It("AlertSettingUnitMgdL is expected", func() {
		Expect(dexcom.AlertSettingUnitMgdL).To(Equal("mg/dL"))
	})

	It("AlertSettingUnitMgdLMinute is expected", func() {
		Expect(dexcom.AlertSettingUnitMgdLMinute).To(Equal("mg/dL/min"))
	})

	It("AlertSettingValueFallMgdLMinuteMaximum is expected", func() {
		Expect(dexcom.AlertSettingValueFallMgdLMinuteMaximum).To(Equal(10.0))
	})

	It("AlertSettingValueFallMgdLMinuteMinimum is expected", func() {
		Expect(dexcom.AlertSettingValueFallMgdLMinuteMinimum).To(Equal(1.0))
	})

	It("AlertSettingValueHighMgdLMaximum is expected", func() {
		Expect(dexcom.AlertSettingValueHighMgdLMaximum).To(Equal(400.0))
	})

	It("AlertSettingValueHighMgdLMinimum is expected", func() {
		Expect(dexcom.AlertSettingValueHighMgdLMinimum).To(Equal(100.0))
	})

	It("AlertSettingValueLowMgdLMaximum is expected", func() {
		Expect(dexcom.AlertSettingValueLowMgdLMaximum).To(Equal(150.0))
	})

	It("AlertSettingValueLowMgdLMinimum is expected", func() {
		Expect(dexcom.AlertSettingValueLowMgdLMinimum).To(Equal(50.0))
	})

	It("AlertSettingValueNoReadingsMgdLMaximum is expected", func() {
		Expect(dexcom.AlertSettingValueNoReadingsMgdLMaximum).To(Equal(360.0))
	})

	It("AlertSettingValueNoReadingsMgdLMinimum is expected", func() {
		Expect(dexcom.AlertSettingValueNoReadingsMgdLMinimum).To(Equal(0.0))
	})

	It("AlertSettingValueOutOfRangeMgdLMaximum is expected", func() {
		Expect(dexcom.AlertSettingValueOutOfRangeMgdLMaximum).To(Equal(360.0))
	})

	It("AlertSettingValueOutOfRangeMgdLMinimum is expected", func() {
		Expect(dexcom.AlertSettingValueOutOfRangeMgdLMinimum).To(Equal(0.0))
	})

	It("AlertSettingValueRiseMgdLMinuteMaximum is expected", func() {
		Expect(dexcom.AlertSettingValueRiseMgdLMinuteMaximum).To(Equal(10.0))
	})

	It("AlertSettingValueRiseMgdLMinuteMinimum is expected", func() {
		Expect(dexcom.AlertSettingValueRiseMgdLMinuteMinimum).To(Equal(1.0))
	})

	It("AlertSettingValueUrgentLowMgdLMaximum is expected", func() {
		Expect(dexcom.AlertSettingValueUrgentLowMgdLMaximum).To(Equal(80.0))
	})

	It("AlertSettingValueUrgentLowMgdLMinimum is expected", func() {
		Expect(dexcom.AlertSettingValueUrgentLowMgdLMinimum).To(Equal(40.0))
	})

	It("AlertSettingValueUrgentLowSoonMgdLMaximum is expected", func() {
		Expect(dexcom.AlertSettingValueUrgentLowSoonMgdLMaximum).To(Equal(80.0))
	})

	It("AlertSettingValueUrgentLowSoonMgdLMinimum is expected", func() {
		Expect(dexcom.AlertSettingValueUrgentLowSoonMgdLMinimum).To(Equal(40.0))
	})

	It("AlertScheduleSettingsDays returns expected", func() {
		Expect(dexcom.AlertScheduleSettingsDays()).To(Equal([]string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}))
	})

	Context("AlertScheduleSettingsDayIndex", func() {
		DescribeTable("return the expected index when the day",
			func(day string, expectedIndex int) {
				Expect(dexcom.AlertScheduleSettingsDayIndex(day)).To(Equal(expectedIndex))
			},
			Entry("is an empty string", "", 0),
			Entry("is sunday", "sunday", 1),
			Entry("is monday", "monday", 2),
			Entry("is tuesday", "tuesday", 3),
			Entry("is wednesday", "wednesday", 4),
			Entry("is thursday", "thursday", 5),
			Entry("is friday", "friday", 6),
			Entry("is saturday", "saturday", 7),
			Entry("is an invalid string", "invalid", 0),
		)
	})

	It("AlertSettingAlertNames returns expected", func() {
		Expect(dexcom.AlertSettingAlertNames()).To(Equal([]string{"fall", "high", "low", "noReadings", "outOfRange", "rise", "urgentLow", "urgentLowSoon"}))
	})

	It("AlertSettingUnitFalls returns expected", func() {
		Expect(dexcom.AlertSettingUnitFalls()).To(Equal([]string{"mg/dL/min"}))
	})

	It("AlertSettingUnitHighs returns expected", func() {
		Expect(dexcom.AlertSettingUnitHighs()).To(Equal([]string{"mg/dL"}))
	})

	It("AlertSettingUnitLows returns expected", func() {
		Expect(dexcom.AlertSettingUnitLows()).To(Equal([]string{"mg/dL"}))
	})

	It("AlertSettingUnitNoReadings returns expected", func() {
		Expect(dexcom.AlertSettingUnitNoReadings()).To(Equal([]string{"minutes"}))
	})

	It("AlertSettingUnitOutOfRanges returns expected", func() {
		Expect(dexcom.AlertSettingUnitOutOfRanges()).To(Equal([]string{"minutes"}))
	})

	It("AlertSettingUnitRises returns expected", func() {
		Expect(dexcom.AlertSettingUnitRises()).To(Equal([]string{"mg/dL/min"}))
	})

	It("AlertSettingUnitUrgentLows returns expected", func() {
		Expect(dexcom.AlertSettingUnitUrgentLows()).To(Equal([]string{"mg/dL"}))
	})

	It("AlertSettingUnitUrgentLowSoons returns expected", func() {
		Expect(dexcom.AlertSettingUnitUrgentLowSoons()).To(Equal([]string{"mg/dL"}))
	})

	Context("ParseAlertScheduleSettingsTime", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedHour int, expectedMinute int, expectedOK bool) {
				hour, minute, ok := dexcom.ParseAlertScheduleSettingsTime(value)
				Expect(ok).To(Equal(expectedOK))
				Expect(hour).To(Equal(expectedHour))
				Expect(minute).To(Equal(expectedMinute))
			},
			Entry("is an empty string", "", 0, 0, false),
			Entry("contains non-numbers", "a$: b", 0, 0, false),
			Entry("does not exactly match format", "1;23", 0, 0, false),
			Entry("has hour in range (lower)", "00:00", 0, 0, true),
			Entry("has hour in range (upper)", "23:59", 23, 59, true),
			Entry("has hour out of range (upper)", "24:00", 0, 0, false),
			Entry("has minute in range (lower)", "00:00", 0, 0, true),
			Entry("has minute in range (upper)", "23:59", 23, 59, true),
			Entry("has minute out of range (upper)", "23:60", 0, 0, false),

			Entry("is 12hr format with AM postfix", "8:00 Am", 8, 0, true),
			Entry("is 12hr format with AM postfix", "08:00 aM", 8, 0, true),
			Entry("is 12hr format with PM postfix", "9:00 Pm", 21, 0, true),
			Entry("is 12hr format with PM postfix and extra padding", "09:00   pm", 21, 0, true),
			Entry("is 12hr format with minutes", "11:59   pM", 23, 59, true),
		)
	})

	Context("IsValidAlertScheduleSettingsTime, AlertScheduleSettingsTimeValidator, and ValidateAlertScheduleSettingsTime", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(dexcom.IsValidAlertScheduleSettingsTime(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				dexcom.AlertScheduleSettingsTimeValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(dexcom.ValidateAlertScheduleSettingsTime(value), expectedErrors...)
			},
			Entry("is an empty string", "", structureValidator.ErrorValueEmpty()),
			Entry("contains non-numbers", "a$: b", dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid("a$: b")),
			Entry("does not exactly match format", "1;23", dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid("1;23")),
			Entry("has hour in range (lower)", "00:00"),
			Entry("has hour in range (upper)", "23:59"),
			Entry("has hour out of range (upper)", "24:00", dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid("24:00")),
			Entry("has minute in range (lower)", "00:00"),
			Entry("has minute in range (upper)", "23:59"),
			Entry("has minute out of range (upper)", "23:60", dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid("23:60")),
		)
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsAlertScheduleSettingsTimeNotValid with empty string", dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as alert schedule settings time`),
			Entry("is ErrorValueStringAsAlertScheduleSettingsTimeNotValid with non-empty string", dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid("XX:XX"), "value-not-valid", "value is not valid", `value "XX:XX" is not valid as alert schedule settings time`),
		)
	})
	Context("AlertSchedules can be", func() {
		DescribeTable("empty",
			func(value *dexcom.AlertSchedules) {
				Expect(value).ToNot(BeNil())
				testValidator := structureValidator.New()
				value.Validate(testValidator)
				Expect(testValidator.HasError()).To(BeFalse())
				Expect(len(*value)).To(Equal(0))
			},
			Entry("valid if empty", dexcom.NewAlertSchedules()),
		)
	})
})
