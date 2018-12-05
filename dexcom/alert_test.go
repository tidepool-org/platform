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

	It("AlertSettingUnitMinutes is expected", func() {
		Expect(dexcom.AlertSettingUnitMinutes).To(Equal("minutes"))
	})

	It("AlertSettingUnitMgdL is expected", func() {
		Expect(dexcom.AlertSettingUnitMgdL).To(Equal("mg/dL"))
	})

	It("AlertSettingUnitMgdLMinute is expected", func() {
		Expect(dexcom.AlertSettingUnitMgdLMinute).To(Equal("mg/dL/min"))
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

	It("AlertSettingValueFallMgdLMinutes returns expected", func() {
		Expect(dexcom.AlertSettingValueFallMgdLMinutes()).To(Equal([]float64{2, 3}))
	})

	It("AlertSettingSnoozeFalls returns expected", func() {
		Expect(dexcom.AlertSettingSnoozeFalls()).To(Equal([]int{0, 30}))
	})

	It("AlertSettingUnitHighs returns expected", func() {
		Expect(dexcom.AlertSettingUnitHighs()).To(Equal([]string{"mg/dL"}))
	})

	It("AlertSettingValueHighMgdLs returns expected", func() {
		Expect(dexcom.AlertSettingValueHighMgdLs()).To(Equal([]float64{120, 130, 140, 150, 160, 170, 180, 190, 200, 210, 220, 230, 240, 250, 260, 270, 280, 290, 300, 310, 320, 330, 340, 350, 360, 370, 380, 390, 400}))
	})

	It("AlertSettingSnoozeHighs returns expected", func() {
		Expect(dexcom.AlertSettingSnoozeHighs()).To(Equal([]int{0, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 85, 90, 95, 100, 105, 110, 115, 120, 125, 130, 135, 140, 145, 150, 155, 160, 165, 170, 175, 180, 185, 190, 195, 200, 205, 210, 215, 220, 225, 230, 235, 240, 255, 270, 285, 300}))
	})

	It("AlertSettingUnitLows returns expected", func() {
		Expect(dexcom.AlertSettingUnitLows()).To(Equal([]string{"mg/dL"}))
	})

	It("AlertSettingValueLowMgdLs returns expected", func() {
		Expect(dexcom.AlertSettingValueLowMgdLs()).To(Equal([]float64{60, 65, 70, 75, 80, 85, 90, 95, 100}))
	})

	It("AlertSettingSnoozeLows returns expected", func() {
		Expect(dexcom.AlertSettingSnoozeLows()).To(Equal([]int{0, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 85, 90, 95, 100, 105, 110, 115, 120, 125, 130, 135, 140, 145, 150, 155, 160, 165, 170, 175, 180, 185, 190, 195, 200, 205, 210, 215, 220, 225, 230, 235, 240, 255, 270, 285, 300}))
	})

	It("AlertSettingUnitNoReadings returns expected", func() {
		Expect(dexcom.AlertSettingUnitNoReadings()).To(Equal([]string{"minutes"}))
	})

	It("AlertSettingValueNoReadingsMinutes returns expected", func() {
		Expect(dexcom.AlertSettingValueNoReadingsMinutes()).To(Equal([]float64{0, 20}))
	})

	It("AlertSettingSnoozeNoReadings returns expected", func() {
		Expect(dexcom.AlertSettingSnoozeNoReadings()).To(Equal([]int{0, 20, 25, 30}))
	})

	It("AlertSettingUnitOutOfRanges returns expected", func() {
		Expect(dexcom.AlertSettingUnitOutOfRanges()).To(Equal([]string{"minutes"}))
	})

	It("AlertSettingValueOutOfRangeMinutes returns expected", func() {
		Expect(dexcom.AlertSettingValueOutOfRangeMinutes()).To(Equal([]float64{20, 25, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 85, 90, 95, 100, 105, 110, 115, 120, 125, 130, 135, 140, 145, 150, 155, 160, 165, 170, 175, 180, 185, 190, 195, 200, 205, 210, 215, 220, 225, 230, 235, 240}))
	})

	It("AlertSettingSnoozeOutOfRanges returns expected", func() {
		Expect(dexcom.AlertSettingSnoozeOutOfRanges()).To(Equal([]int{0, 20, 25, 30}))
	})

	It("AlertSettingUnitRises returns expected", func() {
		Expect(dexcom.AlertSettingUnitRises()).To(Equal([]string{"mg/dL/min"}))
	})

	It("AlertSettingValueRiseMgdLMinutes returns expected", func() {
		Expect(dexcom.AlertSettingValueRiseMgdLMinutes()).To(Equal([]float64{2, 3}))
	})

	It("AlertSettingSnoozeRises returns expected", func() {
		Expect(dexcom.AlertSettingSnoozeRises()).To(Equal([]int{0, 30}))
	})

	It("AlertSettingUnitUrgentLows returns expected", func() {
		Expect(dexcom.AlertSettingUnitUrgentLows()).To(Equal([]string{"mg/dL"}))
	})

	It("AlertSettingValueUrgentLowMgdLs returns expected", func() {
		Expect(dexcom.AlertSettingValueUrgentLowMgdLs()).To(Equal([]float64{55}))
	})

	It("AlertSettingSnoozeUrgentLows returns expected", func() {
		Expect(dexcom.AlertSettingSnoozeUrgentLows()).To(Equal([]int{0, 30}))
	})

	It("AlertSettingUnitUrgentLowSoons returns expected", func() {
		Expect(dexcom.AlertSettingUnitUrgentLowSoons()).To(Equal([]string{"mg/dL"}))
	})

	It("AlertSettingValueUrgentLowSoonMgdLs returns expected", func() {
		Expect(dexcom.AlertSettingValueUrgentLowSoonMgdLs()).To(Equal([]float64{55}))
	})

	It("AlertSettingSnoozeUrgentLowSoons returns expected", func() {
		Expect(dexcom.AlertSettingSnoozeUrgentLowSoons()).To(Equal([]int{0, 30}))
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
			Entry("does not exactly match format", "1:23", 0, 0, false),
			Entry("has hour in range (lower)", "00:00", 0, 0, true),
			Entry("has hour in range (upper)", "23:59", 23, 59, true),
			Entry("has hour out of range (upper)", "24:00", 0, 0, false),
			Entry("has minute in range (lower)", "00:00", 0, 0, true),
			Entry("has minute in range (upper)", "23:59", 23, 59, true),
			Entry("has minute out of range (upper)", "23:60", 0, 0, false),
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
			Entry("does not exactly match format", "1:23", dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid("1:23")),
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
})
