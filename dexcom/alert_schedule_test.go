package dexcom_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/dexcom/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
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
		Expect(dexcom.AlertScheduleSettingsDays()).To(Equal([]string{
			dexcom.AlertScheduleSettingsDaySunday,
			dexcom.AlertScheduleSettingsDayMonday,
			dexcom.AlertScheduleSettingsDayTuesday,
			dexcom.AlertScheduleSettingsDayWednesday,
			dexcom.AlertScheduleSettingsDayThursday,
			dexcom.AlertScheduleSettingsDayFriday,
			dexcom.AlertScheduleSettingsDaySaturday,
		}))
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
		Expect(dexcom.AlertSettingAlertNames()).To(Equal([]string{"unknown", "fall", "high", "low", "noReadings", "outOfRange", "rise", "urgentLow", "urgentLowSoon", "fixedLow"}))
		Expect(dexcom.AlertSettingAlertNames()).To(Equal([]string{
			dexcom.AlertSettingAlertNameUnknown,
			dexcom.AlertSettingAlertNameFall,
			dexcom.AlertSettingAlertNameHigh,
			dexcom.AlertSettingAlertNameLow,
			dexcom.AlertSettingAlertNameNoReadings,
			dexcom.AlertSettingAlertNameOutOfRange,
			dexcom.AlertSettingAlertNameRise,
			dexcom.AlertSettingAlertNameUrgentLow,
			dexcom.AlertSettingAlertNameUrgentLowSoon,
			dexcom.AlertSettingAlertNameFixedLow,
		}))
	})

	It("AlertSettingSoundThemes returns expected", func() {
		Expect(dexcom.AlertSettingSoundThemes()).To(Equal([]string{"unknown", "modern", "classic"}))
		Expect(dexcom.AlertSettingSoundThemes()).To(Equal([]string{
			dexcom.AlertSettingSoundThemeUnknown,
			dexcom.AlertSettingSoundThemeModern,
			dexcom.AlertSettingSoundThemeClassic,
		}))
	})

	It("AlertSettingSoundOutputModes returns expected", func() {
		Expect(dexcom.AlertSettingSoundOutputModes()).To(Equal([]string{"unknown", "sound", "vibrate", "match"}))
		Expect(dexcom.AlertSettingSoundOutputModes()).To(Equal([]string{
			dexcom.AlertSettingSoundOutputModeUnknown,
			dexcom.AlertSettingSoundOutputModeSound,
			dexcom.AlertSettingSoundOutputModeVibrate,
			dexcom.AlertSettingSoundOutputModeMatch,
		}))
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
		DescribeTable("AlertScheduleSettings",
			func(setupFunc func() *dexcom.AlertScheduleSettings) {
				val := setupFunc()
				testValidator := structureValidator.New()
				val.Validate(testValidator)
				Expect(testValidator.HasError()).To(BeFalse())
			},
			Entry("valid if active not set when default", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(true)
				settings.Active = nil
				return settings
			}),
			Entry("valid if active not set", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(false)
				settings.Active = nil
				return settings
			}),
			Entry("valid if override not set", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(false)
				settings.Override = nil
				return settings
			}),
			Entry("valid if override not set when default", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(true)
				settings.Override = nil
				return settings
			}),
			Entry("valid if override.Enabled not set when default", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(true)
				settings.Override = &dexcom.OverrideSetting{
					Enabled: nil,
					Mode:    pointer.FromString(dexcom.AlertScheduleSettingsOverrideModeQuiet),
					EndTime: pointer.FromString(dexcom.AlertScheduleSettingsEndTimeDefault),
				}
				return settings
			}),
			Entry("valid if override.Enabled not set", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(false)
				settings.Override = &dexcom.OverrideSetting{
					Enabled: nil,
					Mode:    pointer.FromString(dexcom.AlertScheduleSettingsOverrideModeQuiet),
					EndTime: pointer.FromString(dexcom.AlertScheduleSettingsEndTimeDefault),
				}
				return settings
			}),
			Entry("valid if override.Mode not set when default", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(true)
				settings.Override = &dexcom.OverrideSetting{
					Enabled: pointer.FromBool(true),
					Mode:    nil,
					EndTime: pointer.FromString(dexcom.AlertScheduleSettingsEndTimeDefault),
				}
				return settings
			}),
			Entry("valid if override.Mode not set", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(false)
				settings.Override = &dexcom.OverrideSetting{
					Enabled: pointer.FromBool(true),
					Mode:    nil,
					EndTime: pointer.FromString(dexcom.AlertScheduleSettingsEndTimeDefault),
				}
				return settings
			}),
			Entry("valid if override.EndTime not set when default", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(true)
				settings.Override = &dexcom.OverrideSetting{
					Enabled: pointer.FromBool(true),
					Mode:    pointer.FromString(dexcom.AlertScheduleSettingsOverrideModeVibrate),
					EndTime: nil,
				}
				return settings
			}),
			Entry("valid if override.EndTime not set", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(false)
				settings.Override = &dexcom.OverrideSetting{
					Enabled: pointer.FromBool(true),
					Mode:    pointer.FromString(dexcom.AlertScheduleSettingsOverrideModeVibrate),
					EndTime: nil,
				}
				return settings
			}),
		)
		DescribeTable("AlertScheduleSettings",
			func(setupFunc func() *dexcom.AlertScheduleSettings, expectError bool) {
				val := setupFunc()
				testValidator := structureValidator.New()
				val.Validate(testValidator)
				Expect(testValidator.HasError()).To(Equal(expectError))
			},
			Entry("valid if active not set when default", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(true)
				settings.Active = nil
				return settings
			}, false),
			Entry("valid if active not set", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(false)
				settings.Active = nil
				return settings
			}, false),
			Entry("valid if override not set", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(false)
				settings.Override = nil
				return settings
			}, false),
			Entry("valid if override not set when default", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(true)
				settings.Override = nil
				return settings
			}, false),
			Entry("valid if override.Enabled not set when default", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(true)
				settings.Override = &dexcom.OverrideSetting{
					Enabled: nil,
					Mode:    pointer.FromString(dexcom.AlertScheduleSettingsOverrideModeQuiet),
					EndTime: pointer.FromString(dexcom.AlertScheduleSettingsEndTimeDefault),
				}
				return settings
			}, false),
			Entry("valid if override.Enabled not set", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(false)
				settings.Override = &dexcom.OverrideSetting{
					Enabled: nil,
					Mode:    pointer.FromString(dexcom.AlertScheduleSettingsOverrideModeQuiet),
					EndTime: pointer.FromString(dexcom.AlertScheduleSettingsEndTimeDefault),
				}
				return settings
			}, false),
			Entry("valid if override.Mode not set when default", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(true)
				settings.Override = &dexcom.OverrideSetting{
					Enabled: pointer.FromBool(true),
					Mode:    nil,
					EndTime: pointer.FromString(dexcom.AlertScheduleSettingsEndTimeDefault),
				}
				return settings
			}, false),
			Entry("valid if override.Mode not set", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(false)
				settings.Override = &dexcom.OverrideSetting{
					Enabled: pointer.FromBool(true),
					Mode:    nil,
					EndTime: pointer.FromString(dexcom.AlertScheduleSettingsEndTimeDefault),
				}
				return settings
			}, false),
			Entry("valid if override.EndTime not set when default", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(true)
				settings.Override = &dexcom.OverrideSetting{
					Enabled: pointer.FromBool(true),
					Mode:    pointer.FromString(dexcom.AlertScheduleSettingsOverrideModeVibrate),
					EndTime: nil,
				}
				return settings
			}, false),
			Entry("valid if override.EndTime not set", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(false)
				settings.Override = &dexcom.OverrideSetting{
					Enabled: pointer.FromBool(true),
					Mode:    pointer.FromString(dexcom.AlertScheduleSettingsOverrideModeVibrate),
					EndTime: nil,
				}
				return settings
			}, false),
			Entry("errors if name not set", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(false)
				settings.Name = nil
				return settings
			}, true),
			Entry("errors if enabled not set", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(false)
				settings.Enabled = nil
				return settings
			}, true),
			Entry("errors if default not set", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(false)
				settings.Default = nil
				return settings
			}, true),
			Entry("errors if startTime not set", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(false)
				settings.StartTime = nil
				return settings
			}, true),
			Entry("errors if endTime not set", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(false)
				settings.EndTime = nil
				return settings
			}, true),
			Entry("errors if daysOfWeek not set", func() *dexcom.AlertScheduleSettings {
				settings := test.RandomAlertScheduleSettings(false)
				settings.DaysOfWeek = nil
				return settings
			}, true),
		)
		DescribeTable("AlertSetting",
			func(setupFunc func() *dexcom.AlertSetting, expectError bool) {
				val := setupFunc()
				testValidator := structureValidator.New()

				val.Validate(testValidator)
				Expect(testValidator.HasError()).To(Equal(expectError))
			},
			Entry("errors if alertName not set", func() *dexcom.AlertSetting {
				settings := test.RandomAlertSetting(nil)
				settings.AlertName = nil
				return settings
			}, true),
			Entry("errors if enabled not set", func() *dexcom.AlertSetting {
				settings := test.RandomAlertSetting(nil)
				settings.Enabled = nil
				return settings
			}, true),
			Entry("errors if name urgentLow has no snooze", func() *dexcom.AlertSetting {
				settings := test.RandomAlertSetting(pointer.FromString(dexcom.AlertSettingAlertNameUrgentLow))
				settings.Snooze = nil
				return settings
			}, true),
			Entry("errors if name urgentLowSoon has no snooze", func() *dexcom.AlertSetting {
				settings := test.RandomAlertSetting(pointer.FromString(dexcom.AlertSettingAlertNameUrgentLowSoon))
				settings.Snooze = nil
				return settings
			}, true),
			Entry("errors if name low has no snooze", func() *dexcom.AlertSetting {
				settings := test.RandomAlertSetting(pointer.FromString(dexcom.AlertSettingAlertNameLow))
				settings.Snooze = nil
				return settings
			}, true),
			Entry("errors if name high has no snooze", func() *dexcom.AlertSetting {
				settings := test.RandomAlertSetting(pointer.FromString(dexcom.AlertSettingAlertNameHigh))
				settings.Snooze = nil
				return settings
			}, true),
			Entry("valid if soundTheme not set", func() *dexcom.AlertSetting {
				settings := test.RandomAlertSetting(nil)
				settings.SoundTheme = nil
				return settings
			}, false),
			Entry("valid if soundOutputMode not set", func() *dexcom.AlertSetting {
				settings := test.RandomAlertSetting(nil)
				settings.SoundOutputMode = nil
				return settings
			}, false),
			Entry("valid name noReadings has no snooze", func() *dexcom.AlertSetting {
				settings := test.RandomAlertSetting(pointer.FromString(dexcom.AlertSettingAlertNameNoReadings))
				settings.Snooze = nil
				return settings
			}, false),
			Entry("valid name noReadings has no delay", func() *dexcom.AlertSetting {
				settings := test.RandomAlertSetting(pointer.FromString(dexcom.AlertSettingAlertNameNoReadings))
				settings.Delay = nil
				return settings
			}, false),
			Entry("valid name rise has no snooze", func() *dexcom.AlertSetting {
				settings := test.RandomAlertSetting(pointer.FromString(dexcom.AlertSettingAlertNameRise))
				settings.Snooze = nil
				return settings
			}, false),
			Entry("valid name outOfRange has no snooze", func() *dexcom.AlertSetting {
				settings := test.RandomAlertSetting(pointer.FromString(dexcom.AlertSettingAlertNameOutOfRange))
				settings.Snooze = nil
				return settings
			}, false),
			Entry("valid name fall has no snooze", func() *dexcom.AlertSetting {
				settings := test.RandomAlertSetting(pointer.FromString(dexcom.AlertSettingAlertNameFall))
				settings.Snooze = nil
				return settings
			}, false),
		)
	})

	It("AlertScheduleSettingsOverrideModes returns expected", func() {
		Expect(dexcom.AlertScheduleSettingsOverrideModes()).To(Equal([]string{"unknown", "quiet", "vibrate"}))
		Expect(dexcom.AlertScheduleSettingsOverrideModes()).To(Equal([]string{dexcom.AlertScheduleSettingsOverrideModeUnknown, dexcom.AlertScheduleSettingsOverrideModeQuiet, dexcom.AlertScheduleSettingsOverrideModeVibrate}))
	})
})
