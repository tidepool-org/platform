package dexcom_test

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	"github.com/tidepool-org/platform/dexcom"
	dexcomTest "github.com/tidepool-org/platform/dexcom/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureNormalizer "github.com/tidepool-org/platform/structure/normalizer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("AlertSchedule", func() {
	It("AlertScheduleSettingsStartTimeDefault returns expected", func() {
		Expect(dexcom.AlertScheduleSettingsStartTimeDefault).To(Equal("00:00"))
	})

	It("AlertScheduleSettingsStartTimeDefaultAlternate returns expected", func() {
		Expect(dexcom.AlertScheduleSettingsStartTimeDefaultAlternate).To(Equal("0:00"))
	})

	It("AlertScheduleSettingsEndTimeDefault returns expected", func() {
		Expect(dexcom.AlertScheduleSettingsEndTimeDefault).To(Equal("00:00"))
	})

	It("AlertScheduleSettingsEndTimeDefaultAlternate returns expected", func() {
		Expect(dexcom.AlertScheduleSettingsEndTimeDefaultAlternate).To(Equal("0:00"))
	})

	It("AlertScheduleSettingsDaySunday returns expected", func() {
		Expect(dexcom.AlertScheduleSettingsDaySunday).To(Equal("sunday"))
	})

	It("AlertScheduleSettingsDayMonday returns expected", func() {
		Expect(dexcom.AlertScheduleSettingsDayMonday).To(Equal("monday"))
	})

	It("AlertScheduleSettingsDayTuesday returns expected", func() {
		Expect(dexcom.AlertScheduleSettingsDayTuesday).To(Equal("tuesday"))
	})

	It("AlertScheduleSettingsDayWednesday returns expected", func() {
		Expect(dexcom.AlertScheduleSettingsDayWednesday).To(Equal("wednesday"))
	})

	It("AlertScheduleSettingsDayThursday returns expected", func() {
		Expect(dexcom.AlertScheduleSettingsDayThursday).To(Equal("thursday"))
	})

	It("AlertScheduleSettingsDayFriday returns expected", func() {
		Expect(dexcom.AlertScheduleSettingsDayFriday).To(Equal("friday"))
	})

	It("AlertScheduleSettingsDaySaturday returns expected", func() {
		Expect(dexcom.AlertScheduleSettingsDaySaturday).To(Equal("saturday"))
	})

	It("AlertScheduleSettingsOverrideModeUnknown returns expected", func() {
		Expect(dexcom.AlertScheduleSettingsOverrideModeUnknown).To(Equal("unknown"))
	})

	It("AlertScheduleSettingsOverrideModeQuiet returns expected", func() {
		Expect(dexcom.AlertScheduleSettingsOverrideModeQuiet).To(Equal("quiet"))
	})

	It("AlertScheduleSettingsOverrideModeVibrate returns expected", func() {
		Expect(dexcom.AlertScheduleSettingsOverrideModeVibrate).To(Equal("vibrate"))
	})

	It("AlertSettingAlertNameUnknown returns expected", func() {
		Expect(dexcom.AlertSettingAlertNameUnknown).To(Equal("unknown"))
	})

	It("AlertSettingAlertNameHigh returns expected", func() {
		Expect(dexcom.AlertSettingAlertNameHigh).To(Equal("high"))
	})

	It("AlertSettingAlertNameLow returns expected", func() {
		Expect(dexcom.AlertSettingAlertNameLow).To(Equal("low"))
	})

	It("AlertSettingAlertNameRise returns expected", func() {
		Expect(dexcom.AlertSettingAlertNameRise).To(Equal("rise"))
	})

	It("AlertSettingAlertNameFall returns expected", func() {
		Expect(dexcom.AlertSettingAlertNameFall).To(Equal("fall"))
	})

	It("AlertSettingAlertNameOutOfRange returns expected", func() {
		Expect(dexcom.AlertSettingAlertNameOutOfRange).To(Equal("outOfRange"))
	})

	It("AlertSettingAlertNameUrgentLow returns expected", func() {
		Expect(dexcom.AlertSettingAlertNameUrgentLow).To(Equal("urgentLow"))
	})

	It("AlertSettingAlertNameUrgentLowSoon returns expected", func() {
		Expect(dexcom.AlertSettingAlertNameUrgentLowSoon).To(Equal("urgentLowSoon"))
	})

	It("AlertSettingAlertNameNoReadings returns expected", func() {
		Expect(dexcom.AlertSettingAlertNameNoReadings).To(Equal("noReadings"))
	})

	It("AlertSettingAlertNameFixedLow returns expected", func() {
		Expect(dexcom.AlertSettingAlertNameFixedLow).To(Equal("fixedLow"))
	})

	It("AlertSettingSnoozeMinutesMaximum returns expected", func() {
		Expect(dexcom.AlertSettingSnoozeMinutesMaximum).To(Equal(dataTypesSettingsCgm.SnoozeDurationMinutesMaximum))
	})

	It("AlertSettingSnoozeMinutesMinimum returns expected", func() {
		Expect(dexcom.AlertSettingSnoozeMinutesMinimum).To(Equal(dataTypesSettingsCgm.SnoozeDurationMinutesMinimum))
	})

	It("AlertSettingDelayMinimum returns expected", func() {
		Expect(dexcom.AlertSettingDelayMinimum).To(Equal(0))
	})

	It("AlertSettingUnitUnknown returns expected", func() {
		Expect(dexcom.AlertSettingUnitUnknown).To(Equal("unknown"))
	})

	It("AlertSettingUnitMinutes returns expected", func() {
		Expect(dexcom.AlertSettingUnitMinutes).To(Equal("minutes"))
	})

	It("AlertSettingUnitMgdL returns expected", func() {
		Expect(dexcom.AlertSettingUnitMgdL).To(Equal("mg/dL"))
	})

	It("AlertSettingUnitMmolL returns expected", func() {
		Expect(dexcom.AlertSettingUnitMmolL).To(Equal("mmol/L"))
	})

	It("AlertSettingUnitMgdLMinute returns expected", func() {
		Expect(dexcom.AlertSettingUnitMgdLMinute).To(Equal("mg/dL/min"))
	})

	It("AlertSettingUnitMmolLMinute returns expected", func() {
		Expect(dexcom.AlertSettingUnitMmolLMinute).To(Equal("mmol/L/min"))
	})

	It("AlertSettingValueHighMgdLMaximum returns expected", func() {
		Expect(dexcom.AlertSettingValueHighMgdLMaximum).To(Equal(dataTypesSettingsCgm.HighAlertLevelMgdLMaximum))
	})

	It("AlertSettingValueHighMgdLMinimum returns expected", func() {
		Expect(dexcom.AlertSettingValueHighMgdLMinimum).To(Equal(dataTypesSettingsCgm.HighAlertLevelMgdLMinimum))
	})

	It("AlertSettingValueHighMmolLMaximum returns expected", func() {
		Expect(dexcom.AlertSettingValueHighMmolLMaximum).To(Equal(dataTypesSettingsCgm.HighAlertLevelMmolLMaximum))
	})

	It("AlertSettingValueHighMmolLMinimum returns expected", func() {
		Expect(dexcom.AlertSettingValueHighMmolLMinimum).To(Equal(dataTypesSettingsCgm.HighAlertLevelMmolLMinimum))
	})

	It("AlertSettingValueLowMgdLMaximum returns expected", func() {
		Expect(dexcom.AlertSettingValueLowMgdLMaximum).To(Equal(dataTypesSettingsCgm.LowAlertLevelMgdLMaximum))
	})

	It("AlertSettingValueLowMgdLMinimum returns expected", func() {
		Expect(dexcom.AlertSettingValueLowMgdLMinimum).To(Equal(dataTypesSettingsCgm.LowAlertLevelMgdLMinimum))
	})

	It("AlertSettingValueLowMmolLMaximum returns expected", func() {
		Expect(dexcom.AlertSettingValueLowMmolLMaximum).To(Equal(dataTypesSettingsCgm.LowAlertLevelMmolLMaximum))
	})

	It("AlertSettingValueLowMmolLMinimum returns expected", func() {
		Expect(dexcom.AlertSettingValueLowMmolLMinimum).To(Equal(dataTypesSettingsCgm.LowAlertLevelMmolLMinimum))
	})

	It("AlertSettingValueRiseMgdLMinuteMaximum returns expected", func() {
		Expect(dexcom.AlertSettingValueRiseMgdLMinuteMaximum).To(Equal(dataTypesSettingsCgm.RiseAlertRateMgdLMinuteMaximum))
	})

	It("AlertSettingValueRiseMgdLMinuteMinimum returns expected", func() {
		Expect(dexcom.AlertSettingValueRiseMgdLMinuteMinimum).To(Equal(dataTypesSettingsCgm.RiseAlertRateMgdLMinuteMinimum))
	})

	It("AlertSettingValueRiseMmolLMinuteMaximum returns expected", func() {
		Expect(dexcom.AlertSettingValueRiseMmolLMinuteMaximum).To(Equal(dataTypesSettingsCgm.RiseAlertRateMmolLMinuteMaximum))
	})

	It("AlertSettingValueRiseMmolLMinuteMinimum returns expected", func() {
		Expect(dexcom.AlertSettingValueRiseMmolLMinuteMinimum).To(Equal(dataTypesSettingsCgm.RiseAlertRateMmolLMinuteMinimum))
	})

	It("AlertSettingValueFallMgdLMinuteMaximum returns expected", func() {
		Expect(dexcom.AlertSettingValueFallMgdLMinuteMaximum).To(Equal(dataTypesSettingsCgm.FallAlertRateMgdLMinuteMaximum))
	})

	It("AlertSettingValueFallMgdLMinuteMinimum returns expected", func() {
		Expect(dexcom.AlertSettingValueFallMgdLMinuteMinimum).To(Equal(dataTypesSettingsCgm.FallAlertRateMgdLMinuteMinimum))
	})

	It("AlertSettingValueFallMmolLMinuteMaximum returns expected", func() {
		Expect(dexcom.AlertSettingValueFallMmolLMinuteMaximum).To(Equal(dataTypesSettingsCgm.FallAlertRateMmolLMinuteMaximum))
	})

	It("AlertSettingValueFallMmolLMinuteMinimum returns expected", func() {
		Expect(dexcom.AlertSettingValueFallMmolLMinuteMinimum).To(Equal(dataTypesSettingsCgm.FallAlertRateMmolLMinuteMinimum))
	})

	It("AlertSettingValueOutOfRangeMinutesMaximum returns expected", func() {
		Expect(dexcom.AlertSettingValueOutOfRangeMinutesMaximum).To(Equal(dataTypesSettingsCgm.OutOfRangeAlertDurationMinutesMaximum))
	})

	It("AlertSettingValueOutOfRangeMinutesMinimum returns expected", func() {
		Expect(dexcom.AlertSettingValueOutOfRangeMinutesMinimum).To(Equal(dataTypesSettingsCgm.OutOfRangeAlertDurationMinutesMinimum))
	})

	It("AlertSettingValueUrgentLowMgdLMaximum returns expected", func() {
		Expect(dexcom.AlertSettingValueUrgentLowMgdLMaximum).To(Equal(dataTypesSettingsCgm.UrgentLowAlertLevelMgdLMaximum))
	})

	It("AlertSettingValueUrgentLowMgdLMinimum returns expected", func() {
		Expect(dexcom.AlertSettingValueUrgentLowMgdLMinimum).To(Equal(dataTypesSettingsCgm.UrgentLowAlertLevelMgdLMinimum))
	})

	It("AlertSettingValueUrgentLowMmolLMaximum returns expected", func() {
		Expect(dexcom.AlertSettingValueUrgentLowMmolLMaximum).To(Equal(dataTypesSettingsCgm.UrgentLowAlertLevelMmolLMaximum))
	})

	It("AlertSettingValueUrgentLowMmolLMinimum returns expected", func() {
		Expect(dexcom.AlertSettingValueUrgentLowMmolLMinimum).To(Equal(dataTypesSettingsCgm.UrgentLowAlertLevelMmolLMinimum))
	})

	It("AlertSettingValueUrgentLowSoonMgdLMaximum returns expected", func() {
		Expect(dexcom.AlertSettingValueUrgentLowSoonMgdLMaximum).To(Equal(dataTypesSettingsCgm.UrgentLowAlertLevelMgdLMaximum))
	})

	It("AlertSettingValueUrgentLowSoonMgdLMinimum returns expected", func() {
		Expect(dexcom.AlertSettingValueUrgentLowSoonMgdLMinimum).To(Equal(dataTypesSettingsCgm.UrgentLowAlertLevelMgdLMinimum))
	})

	It("AlertSettingValueUrgentLowSoonMmolLMaximum returns expected", func() {
		Expect(dexcom.AlertSettingValueUrgentLowSoonMmolLMaximum).To(Equal(dataTypesSettingsCgm.UrgentLowAlertLevelMmolLMaximum))
	})

	It("AlertSettingValueUrgentLowSoonMmolLMinimum returns expected", func() {
		Expect(dexcom.AlertSettingValueUrgentLowSoonMmolLMinimum).To(Equal(dataTypesSettingsCgm.UrgentLowAlertLevelMmolLMinimum))
	})

	It("AlertSettingValueNoReadingsMinutesMaximum returns expected", func() {
		Expect(dexcom.AlertSettingValueNoReadingsMinutesMaximum).To(Equal(dataTypesSettingsCgm.NoDataAlertDurationMinutesMaximum))
	})

	It("AlertSettingValueNoReadingsMinutesMinimum returns expected", func() {
		Expect(dexcom.AlertSettingValueNoReadingsMinutesMinimum).To(Equal(dataTypesSettingsCgm.NoDataAlertDurationMinutesMinimum))
	})

	It("AlertSettingValueFixedLowMinutesMaximum returns expected", func() {
		Expect(dexcom.AlertSettingValueFixedLowMinutesMaximum).To(Equal(dataTypesSettingsCgm.NoDataAlertDurationMinutesMaximum))
	})

	It("AlertSettingValueFixedLowMinutesMinimum returns expected", func() {
		Expect(dexcom.AlertSettingValueFixedLowMinutesMinimum).To(Equal(dataTypesSettingsCgm.NoDataAlertDurationMinutesMinimum))
	})

	It("AlertSettingSoundThemeUnknown returns expected", func() {
		Expect(dexcom.AlertSettingSoundThemeUnknown).To(Equal("unknown"))
	})

	It("AlertSettingSoundThemeModern returns expected", func() {
		Expect(dexcom.AlertSettingSoundThemeModern).To(Equal("modern"))
	})

	It("AlertSettingSoundThemeClassic returns expected", func() {
		Expect(dexcom.AlertSettingSoundThemeClassic).To(Equal("classic"))
	})

	It("AlertSettingSoundOutputModeUnknown returns expected", func() {
		Expect(dexcom.AlertSettingSoundOutputModeUnknown).To(Equal("unknown"))
	})

	It("AlertSettingSoundOutputModeSound returns expected", func() {
		Expect(dexcom.AlertSettingSoundOutputModeSound).To(Equal("sound"))
	})

	It("AlertSettingSoundOutputModeVibrate returns expected", func() {
		Expect(dexcom.AlertSettingSoundOutputModeVibrate).To(Equal("vibrate"))
	})

	It("AlertSettingSoundOutputModeMatch returns expected", func() {
		Expect(dexcom.AlertSettingSoundOutputModeMatch).To(Equal("match"))
	})

	It("AlertScheduleSettingsDays returns expected", func() {
		Expect(dexcom.AlertScheduleSettingsDays()).To(Equal([]string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}))
	})

	It("AlertScheduleSettingsOverrideModes returns expected", func() {
		Expect(dexcom.AlertScheduleSettingsOverrideModes()).To(Equal([]string{"unknown", "quiet", "vibrate"}))
	})

	It("AlertSettingAlertNames returns expected", func() {
		Expect(dexcom.AlertSettingAlertNames()).To(Equal([]string{"unknown", "high", "low", "rise", "fall", "outOfRange", "urgentLow", "urgentLowSoon", "noReadings", "fixedLow"}))
	})

	It("AlertSettingSoundThemes returns expected", func() {
		Expect(dexcom.AlertSettingSoundThemes()).To(Equal([]string{"unknown", "modern", "classic"}))
	})

	It("AlertSettingSoundOutputModes returns expected", func() {
		Expect(dexcom.AlertSettingSoundOutputModes()).To(Equal([]string{"unknown", "sound", "vibrate", "match"}))
	})

	It("AlertSettingUnitUnknowns returns expected", func() {
		Expect(dexcom.AlertSettingUnitUnknowns()).To(Equal([]string{"unknown"}))
	})

	It("AlertSettingUnitHighs returns expected", func() {
		Expect(dexcom.AlertSettingUnitHighs()).To(Equal([]string{"mg/dL", "mmol/L"}))
	})

	It("AlertSettingUnitLows returns expected", func() {
		Expect(dexcom.AlertSettingUnitLows()).To(Equal([]string{"mg/dL", "mmol/L"}))
	})

	It("AlertSettingUnitRises returns expected", func() {
		Expect(dexcom.AlertSettingUnitRises()).To(Equal([]string{"mg/dL/min", "mmol/L/min"}))
	})

	It("AlertSettingUnitFalls returns expected", func() {
		Expect(dexcom.AlertSettingUnitFalls()).To(Equal([]string{"mg/dL/min", "mmol/L/min"}))
	})

	It("AlertSettingUnitOutOfRanges returns expected", func() {
		Expect(dexcom.AlertSettingUnitOutOfRanges()).To(Equal([]string{"minutes"}))
	})

	It("AlertSettingUnitUrgentLows returns expected", func() {
		Expect(dexcom.AlertSettingUnitUrgentLows()).To(Equal([]string{"mg/dL", "mmol/L"}))
	})

	It("AlertSettingUnitUrgentLowSoons returns expected", func() {
		Expect(dexcom.AlertSettingUnitUrgentLowSoons()).To(Equal([]string{"mg/dL", "mmol/L"}))
	})

	It("AlertSettingUnitNoReadings returns expected", func() {
		Expect(dexcom.AlertSettingUnitNoReadings()).To(Equal([]string{"minutes"}))
	})

	It("AlertSettingUnitFixedLows returns expected", func() {
		Expect(dexcom.AlertSettingUnitFixedLows()).To(Equal([]string{"minutes"}))
	})

	Context("AlertSchedulesByName", func() {
		It("sorts alertSchedules with missing in front", func() {
			for _, length := range []int{0, 1, 2, 3, 5, 10} {
				alertSchedulesNil := make(dexcom.AlertSchedules, length)
				alertSchedulesWithoutName := make(dexcom.AlertSchedules, length)
				for index := range alertSchedulesWithoutName {
					alertSchedule := dexcomTest.RandomAlertScheduleWithDefault(false)
					alertSchedule.AlertScheduleSettings.AlertScheduleName = nil
					alertSchedulesWithoutName[index] = alertSchedule
				}
				alertSchedulesWithName := make(dexcom.AlertSchedules, length)
				for index := range alertSchedulesWithName {
					alertSchedule := dexcomTest.RandomAlertScheduleWithDefault(false)
					alertSchedule.AlertScheduleSettings.AlertScheduleName = pointer.FromString(strconv.Itoa(index))
					alertSchedulesWithName[index] = alertSchedule
				}

				expectedAlertSchedules := append(alertSchedulesNil, append(alertSchedulesWithoutName, alertSchedulesWithName...)...)
				actualAlertSchedules := make(dexcom.AlertSchedules, len(expectedAlertSchedules))
				for index, value := range rand.Perm(len(actualAlertSchedules)) {
					actualAlertSchedules[value] = expectedAlertSchedules[index]
				}

				// Fix order of expected alert schedules without name (after randomizing)
				withoutNameIndex := length
				for _, alertSchedule := range actualAlertSchedules {
					if alertSchedule != nil && alertSchedule.Name() == nil {
						expectedAlertSchedules[withoutNameIndex] = alertSchedule
						withoutNameIndex += 1
					}
				}

				sort.Stable(dexcom.AlertSchedulesByName(actualAlertSchedules))
				Expect(actualAlertSchedules).To(Equal(expectedAlertSchedules))
			}
		})
	})

	Context("ParseAlertSchedule", func() {
		It("returns nil if the object is nil", func() {
			parser := structureParser.NewObject(logTest.NewLogger(), nil)
			Expect(dexcom.ParseAlertSchedule(parser)).To(BeNil())
		})

		It("returns the parsed object", func() {
			expectedDatum := dexcomTest.RandomAlertSchedule()
			object := dexcomTest.NewObjectFromAlertSchedule(expectedDatum, test.ObjectFormatJSON)
			parser := structureParser.NewObject(logTest.NewLogger(), &object)
			Expect(dexcom.ParseAlertSchedule(parser)).To(Equal(expectedDatum))
		})
	})

	Context("AlertSchedule", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dexcom.AlertSchedule), expectedErrors ...error) {
					expectedDatum := dexcomTest.RandomAlertSchedule()
					object := dexcomTest.NewObjectFromAlertSchedule(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &dexcom.AlertSchedule{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertSchedule) {},
				),
				Entry("alertScheduleSettings invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertSchedule) {
						object["alertScheduleSettings"] = true
						expectedDatum.AlertScheduleSettings = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/alertScheduleSettings"),
				),
				Entry("alertSettings invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertSchedule) {
						object["alertSettings"] = true
						expectedDatum.AlertSettings = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/alertSettings"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dexcom.AlertSchedule), expectedErrors ...error) {
					datum := dexcomTest.RandomAlertSchedule()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.AlertSchedule) {},
				),
				Entry("alertScheduleSettings missing",
					func(datum *dexcom.AlertSchedule) {
						datum.AlertScheduleSettings = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alertScheduleSettings"),
				),
				Entry("alertScheduleSettings invalid",
					func(datum *dexcom.AlertSchedule) {
						*datum = *dexcomTest.RandomAlertScheduleWithDefault(false)
						datum.AlertScheduleSettings.IsDefaultSchedule = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alertScheduleSettings/isDefaultSchedule"),
				),
				Entry("alertSettings missing",
					func(datum *dexcom.AlertSchedule) {
						datum.AlertSettings = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alertSettings"),
				),
				Entry("alertSettings invalid",
					func(datum *dexcom.AlertSchedule) {
						(*datum.AlertSettings)[0].SystemTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alertSettings/0/systemTime"),
				),
				Entry("multiple errors",
					func(datum *dexcom.AlertSchedule) {
						datum.AlertScheduleSettings = nil
						datum.AlertSettings = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alertScheduleSettings"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alertSettings"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *dexcom.AlertSchedule), expectator func(datum *dexcom.AlertSchedule, expectedDatum *dexcom.AlertSchedule)) {
					for _, origin := range structure.Origins() {
						datum := dexcomTest.RandomAlertSchedule()
						mutator(datum)
						expectedDatum := dexcomTest.CloneAlertSchedule(datum)
						normalizer := structureNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						if expectator != nil {
							expectator(datum, expectedDatum)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("normalizes the alertSchedule",
					func(datum *dexcom.AlertSchedule) {},
					func(datum *dexcom.AlertSchedule, expectedDatum *dexcom.AlertSchedule) {
						normalizer := structureNormalizer.New(logTest.NewLogger())
						expectedDatum.AlertScheduleSettings.Normalize(normalizer)
						expectedDatum.AlertSettings.Normalize(normalizer)
					},
				),
			)
		})

		Context("IsDefault", func() {
			It("returns false if alertScheduleSettings is nil", func() {
				datum := dexcomTest.RandomAlertSchedule()
				datum.AlertScheduleSettings = nil
				Expect(datum.IsDefault()).To(BeFalse())
			})

			It("returns false if alertScheduleSettings is not default", func() {
				datum := dexcomTest.RandomAlertSchedule()
				datum.AlertScheduleSettings.IsDefaultSchedule = pointer.FromBool(false)
				Expect(datum.IsDefault()).To(BeFalse())
			})

			It("returns true if alertScheduleSettings is default", func() {
				datum := dexcomTest.RandomAlertSchedule()
				datum.AlertScheduleSettings.IsDefaultSchedule = pointer.FromBool(true)
				Expect(datum.IsDefault()).To(BeTrue())
			})
		})

		Context("Name", func() {
			It("returns nil is alertScheduleSettings is nil", func() {
				datum := dexcomTest.RandomAlertSchedule()
				datum.AlertScheduleSettings = nil
				Expect(datum.Name()).To(BeNil())
			})

			It("returns alertScheduleSettings name", func() {
				name := pointer.FromString(test.RandomString())
				datum := dexcomTest.RandomAlertSchedule()
				datum.AlertScheduleSettings.AlertScheduleName = name
				Expect(datum.Name()).To(Equal(name))
			})
		})
	})

	Context("ParseAlertScheduleSettings", func() {
		It("returns nil if the object is nil", func() {
			parser := structureParser.NewObject(logTest.NewLogger(), nil)
			Expect(dexcom.ParseAlertScheduleSettings(parser)).To(BeNil())
		})

		It("returns the parsed object", func() {
			expectedDatum := dexcomTest.RandomAlertScheduleSettings()
			object := dexcomTest.NewObjectFromAlertScheduleSettings(expectedDatum, test.ObjectFormatJSON)
			parser := structureParser.NewObject(logTest.NewLogger(), &object)
			Expect(dexcom.ParseAlertScheduleSettings(parser)).To(Equal(expectedDatum))
		})
	})

	Context("AlertScheduleSettings", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dexcom.AlertScheduleSettings), expectedErrors ...error) {
					expectedDatum := dexcomTest.RandomAlertScheduleSettings()
					object := dexcomTest.NewObjectFromAlertScheduleSettings(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &dexcom.AlertScheduleSettings{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertScheduleSettings) {},
				),
				Entry("isDefaultSchedule invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertScheduleSettings) {
						object["isDefaultSchedule"] = ""
						expectedDatum.IsDefaultSchedule = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool(""), "/isDefaultSchedule"),
				),
				Entry("isEnabled invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertScheduleSettings) {
						object["isEnabled"] = ""
						expectedDatum.IsEnabled = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool(""), "/isEnabled"),
				),
				Entry("isActive invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertScheduleSettings) {
						object["isActive"] = ""
						expectedDatum.IsActive = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool(""), "/isActive"),
				),
				Entry("alertScheduleName invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertScheduleSettings) {
						object["alertScheduleName"] = true
						expectedDatum.AlertScheduleName = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/alertScheduleName"),
				),
				Entry("startTime invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertScheduleSettings) {
						object["startTime"] = true
						expectedDatum.StartTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/startTime"),
				),
				Entry("startTime invalid time",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertScheduleSettings) {
						object["startTime"] = "invalid"
						expectedDatum.StartTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", "15:04"), "/startTime"),
				),
				Entry("endTime invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertScheduleSettings) {
						object["endTime"] = true
						expectedDatum.EndTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/endTime"),
				),
				Entry("endTime invalid time",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertScheduleSettings) {
						object["endTime"] = "invalid"
						expectedDatum.EndTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", "15:04"), "/endTime"),
				),
				Entry("daysOfWeek invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertScheduleSettings) {
						object["daysOfWeek"] = true
						expectedDatum.DaysOfWeek = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/daysOfWeek"),
				),
				Entry("override invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertScheduleSettings) {
						object["override"] = true
						expectedDatum.Override = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/override"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(isDefault bool, mutator func(datum *dexcom.AlertScheduleSettings), expectedErrors ...error) {
					datum := dexcomTest.RandomAlertScheduleSettingsWithDefault(isDefault)
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					test.RandomBool(),
					func(datum *dexcom.AlertScheduleSettings) {},
				),
				Entry("isDefaultSchedule missing",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.IsDefaultSchedule = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/isDefaultSchedule"),
				),
				Entry("isDefaultSchedule is true; isEnabled missing",
					true,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.IsEnabled = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/isEnabled"),
				),
				Entry("isDefaultSchedule is true; alertScheduleName missing",
					true,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.AlertScheduleName = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alertScheduleName"),
				),
				Entry("isDefaultSchedule is true; alertScheduleName present",
					true,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.AlertScheduleName = pointer.FromString("present")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEmpty(), "/alertScheduleName"),
				),
				Entry("isDefaultSchedule is true; startTime missing, forced to default",
					true,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.StartTime = nil
					},
				),
				Entry("isDefaultSchedule is true; startTime invalid, forced to default",
					true,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.StartTime = pointer.FromString("invalid")
					},
				),
				Entry("isDefaultSchedule is true; startTime not default, forced to default",
					true,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.StartTime = pointer.FromString("00:01")
					},
				),
				Entry("isDefaultSchedule is true; endTime missing, forced to default",
					true,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.EndTime = nil
					},
				),
				Entry("isDefaultSchedule is true; endTime invalid, forced to default",
					true,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.EndTime = pointer.FromString("invalid")
					},
				),
				Entry("isDefaultSchedule is true; endTime not default, forced to default",
					true,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.EndTime = pointer.FromString("23:59")
					},
				),
				Entry("isDefaultSchedule is true; daysOfWeek missing",
					true,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.DaysOfWeek = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/daysOfWeek"),
				),
				Entry("isDefaultSchedule is true; daysOfWeek contains invalid",
					true,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.DaysOfWeek = pointer.FromStringArray([]string{"friday", "invalid", "monday"})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.AlertScheduleSettingsDays()), "/daysOfWeek/1"),
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotEqualTo(3, len(dexcom.AlertScheduleSettingsDays())), "/daysOfWeek"),
				),
				Entry("isDefaultSchedule is true; daysOfWeek contains duplicates",
					true,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.DaysOfWeek = pointer.FromStringArray([]string{"friday", "monday", "friday"})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/daysOfWeek/2"),
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotEqualTo(3, len(dexcom.AlertScheduleSettingsDays())), "/daysOfWeek"),
				),
				Entry("isDefaultSchedule is true; daysOfWeek does not contain all days",
					true,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.DaysOfWeek = pointer.FromStringArray([]string{"friday", "monday", "saturday", "tuesday", "thursday", "sunday"})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotEqualTo(6, len(dexcom.AlertScheduleSettingsDays())), "/daysOfWeek"),
				),
				Entry("isDefaultSchedule is true; daysOfWeek does not contain all days",
					true,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.DaysOfWeek = pointer.FromStringArray([]string{"friday", "monday", "saturday", "", "tuesday", "thursday", "sunday", "wednesday"})
					},
				),
				Entry("isDefaultSchedule is true; override invalid",
					true,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.Override = dexcomTest.RandomOverride()
						datum.Override.IsOverrideEnabled = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/override/isOverrideEnabled"),
				),
				Entry("isDefaultSchedule is false; isEnabled missing",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.IsEnabled = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/isEnabled"),
				),
				Entry("isDefaultSchedule is false; alertScheduleName missing",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.AlertScheduleName = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alertScheduleName"),
				),
				Entry("isDefaultSchedule is false; alertScheduleName empty",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.AlertScheduleName = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/alertScheduleName"),
				),
				Entry("isDefaultSchedule is false; startTime missing",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.StartTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/startTime"),
				),
				Entry("isDefaultSchedule is false; startTime empty",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.StartTime = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/startTime"),
				),
				Entry("isDefaultSchedule is false; startTime invalid",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.StartTime = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid("invalid"), "/startTime"),
				),
				Entry("isDefaultSchedule is false; startTime invalid hour",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.StartTime = pointer.FromString("24:00")
					},
					errorsTest.WithPointerSource(dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid("24:00"), "/startTime"),
				),
				Entry("isDefaultSchedule is false; startTime invalid minute",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.StartTime = pointer.FromString("23:60")
					},
					errorsTest.WithPointerSource(dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid("23:60"), "/startTime"),
				),
				Entry("isDefaultSchedule is false; endTime missing",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.EndTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/endTime"),
				),
				Entry("isDefaultSchedule is false; endTime empty",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.EndTime = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/endTime"),
				),
				Entry("isDefaultSchedule is false; endTime invalid",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.EndTime = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid("invalid"), "/endTime"),
				),
				Entry("isDefaultSchedule is false; endTime invalid hour",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.EndTime = pointer.FromString("48:00")
					},
					errorsTest.WithPointerSource(dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid("48:00"), "/endTime"),
				),
				Entry("isDefaultSchedule is false; endTime invalid minute",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.EndTime = pointer.FromString("47:60")
					},
					errorsTest.WithPointerSource(dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid("47:60"), "/endTime"),
				),
				Entry("isDefaultSchedule is false; daysOfWeek missing",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.DaysOfWeek = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/daysOfWeek"),
				),
				Entry("isDefaultSchedule is false; daysOfWeek contains invalid",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.DaysOfWeek = pointer.FromStringArray([]string{"friday", "invalid", "monday"})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.AlertScheduleSettingsDays()), "/daysOfWeek/1"),
				),
				Entry("isDefaultSchedule is false; daysOfWeek contains duplicates",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.DaysOfWeek = pointer.FromStringArray([]string{"friday", "monday", "friday"})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/daysOfWeek/2"),
				),
				Entry("isDefaultSchedule is false; daysOfWeek contains empty",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.DaysOfWeek = pointer.FromStringArray([]string{"friday", "", "monday"})
					},
				),
				Entry("isDefaultSchedule is false; override invalid",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.Override = dexcomTest.RandomOverride()
						datum.Override.IsOverrideEnabled = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/override/isOverrideEnabled"),
				),
				Entry("multiple errors",
					false,
					func(datum *dexcom.AlertScheduleSettings) {
						datum.IsDefaultSchedule = nil
						datum.IsEnabled = nil
						datum.AlertScheduleName = nil
						datum.StartTime = nil
						datum.EndTime = nil
						datum.DaysOfWeek = nil
						datum.Override = dexcomTest.RandomOverride()
						datum.Override.IsOverrideEnabled = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/isDefaultSchedule"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/isEnabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alertScheduleName"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/startTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/endTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/daysOfWeek"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/override/isOverrideEnabled"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *dexcom.AlertScheduleSettings), expectator func(datum *dexcom.AlertScheduleSettings, expectedDatum *dexcom.AlertScheduleSettings)) {
					for _, origin := range structure.Origins() {
						datum := dexcomTest.RandomAlertScheduleSettings()
						mutator(datum)
						expectedDatum := dexcomTest.CloneAlertScheduleSettings(datum)
						normalizer := structureNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						if expectator != nil {
							expectator(datum, expectedDatum)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *dexcom.AlertScheduleSettings) {
						datum.DaysOfWeek = nil
					},
					nil,
				),
				Entry("normalizes the daysOfWeek",
					func(datum *dexcom.AlertScheduleSettings) {},
					func(datum *dexcom.AlertScheduleSettings, expectedDatum *dexcom.AlertScheduleSettings) {
						sort.Sort(dexcom.DaysOfWeekByDay(*expectedDatum.DaysOfWeek))
					},
				),
			)
		})

		Context("IsDefault", func() {
			It("returns false if isDefaultSchedule is nil", func() {
				datum := dexcomTest.RandomAlertScheduleSettings()
				datum.IsDefaultSchedule = nil
				Expect(datum.IsDefault()).To(BeFalse())
			})

			It("returns false if isDefaultSchedule is false", func() {
				datum := dexcomTest.RandomAlertScheduleSettings()
				datum.IsDefaultSchedule = pointer.FromBool(false)
				Expect(datum.IsDefault()).To(BeFalse())
			})

			It("returns true if isDefaultSchedule is true", func() {
				datum := dexcomTest.RandomAlertScheduleSettings()
				datum.IsDefaultSchedule = pointer.FromBool(true)
				Expect(datum.IsDefault()).To(BeTrue())
			})
		})
	})

	Context("ParseOverride", func() {
		It("returns nil if the object is nil", func() {
			parser := structureParser.NewObject(logTest.NewLogger(), nil)
			Expect(dexcom.ParseOverride(parser)).To(BeNil())
		})

		It("returns the parsed object", func() {
			expectedDatum := dexcomTest.RandomOverride()
			object := dexcomTest.NewObjectFromOverride(expectedDatum, test.ObjectFormatJSON)
			parser := structureParser.NewObject(logTest.NewLogger(), &object)
			Expect(dexcom.ParseOverride(parser)).To(Equal(expectedDatum))
		})
	})

	Context("Override", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dexcom.Override), expectedErrors ...error) {
					expectedDatum := dexcomTest.RandomOverride()
					object := dexcomTest.NewObjectFromOverride(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &dexcom.Override{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dexcom.Override) {},
				),
				Entry("isOverrideEnabled invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Override) {
						object["isOverrideEnabled"] = ""
						expectedDatum.IsOverrideEnabled = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool(""), "/isOverrideEnabled"),
				),
				Entry("mode invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Override) {
						object["mode"] = true
						expectedDatum.Mode = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/mode"),
				),
				Entry("endTime invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Override) {
						object["endTime"] = true
						expectedDatum.EndTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/endTime"),
				),
				Entry("endTime invalid time",
					func(object map[string]interface{}, expectedDatum *dexcom.Override) {
						object["endTime"] = "invalid"
						expectedDatum.EndTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/endTime"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dexcom.Override), expectedErrors ...error) {
					datum := dexcomTest.RandomOverride()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.Override) {},
				),
				Entry("isDefaultSchedule missing",
					func(datum *dexcom.Override) {
						datum.IsOverrideEnabled = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/isOverrideEnabled"),
				),
				Entry("mode missing",
					func(datum *dexcom.Override) {
						datum.Mode = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mode"),
				),
				Entry("mode invalid",
					func(datum *dexcom.Override) {
						datum.Mode = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.AlertScheduleSettingsOverrideModes()), "/mode"),
				),
				Entry("endTime missing",
					func(datum *dexcom.Override) {
						datum.EndTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/endTime"),
				),
				Entry("endTime zero",
					func(datum *dexcom.Override) {
						datum.EndTime.Time = time.Time{}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/endTime"),
				),
				Entry("multiple errors",
					func(datum *dexcom.Override) {
						datum.IsOverrideEnabled = nil
						datum.Mode = nil
						datum.EndTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/isOverrideEnabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mode"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/endTime"),
				),
			)
		})
	})

	Context("DaysOfWeekByDay", func() {
		DescribeTable("sorts daysOfWeek by day when it",
			func(value []string, expectedValue []string) {
				sort.Sort(dexcom.DaysOfWeekByDay(value))
				Expect(value).To(Equal(expectedValue))
			},
			Entry("is empty", []string{}, []string{}),
			Entry("has single element", []string{"monday"}, []string{"monday"}),
			Entry("has single empty element", []string{""}, []string{""}),
			Entry("has single invalid element", []string{"invalid"}, []string{"invalid"}),
			Entry("has multiple elements", []string{"thursday", "monday", "saturday", "wednesday"}, []string{"monday", "wednesday", "thursday", "saturday"}),
			Entry("has multiple elements with extras", []string{"thursday", "", "2", "monday", "saturday", "1", "wednesday"}, []string{"monday", "wednesday", "thursday", "saturday", "", "1", "2"}),
			Entry("has all elements", []string{"thursday", "sunday", "monday", "saturday", "friday", "wednesday", "tuesday"}, []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}),
			Entry("has all elements with extras", []string{"", "thursday", "2", "sunday", "monday", "saturday", "1", "friday", "wednesday", "tuesday"}, []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "", "1", "2"}),
		)
	})

	Context("ParseAlertScheduleSettingsTime", func() {
		It("returns nil if the referenced value is not found", func() {
			parser := structureParser.NewObject(logTest.NewLogger(), &map[string]any{})
			Expect(dexcom.ParseAlertScheduleSettingsTime(parser, "value")).To(BeNil())
			Expect(parser.Error()).ToNot(HaveOccurred())
		})

		DescribeTable("return nil and reports an error when the referenced value",
			func(value string) {
				parser := structureParser.NewObject(logTest.NewLogger(), &map[string]any{"value": value})
				Expect(dexcom.ParseAlertScheduleSettingsTime(parser, "value")).To(BeNil())
				Expect(parser.Error()).To(MatchError(fmt.Sprintf(`value "%s" is not a parsable time of format "15:04"`, value)))
			},
			Entry("is an empty string", ""),
			Entry("contains non-numbers", "a$: b"),
			Entry("does not exactly match format", "1;23"),
		)

		DescribeTable("return the expected results and no error when the referenced value",
			func(value string, expectedString string) {
				parser := structureParser.NewObject(logTest.NewLogger(), &map[string]any{"value": value})
				Expect(dexcom.ParseAlertScheduleSettingsTime(parser, "value")).To(Equal(&expectedString))
				Expect(parser.Error()).ToNot(HaveOccurred())
			},
			Entry("has hour and minute minimum", "00:00", "00:00"),
			Entry("has hour and minute maximum", "99:99", "99:99"),
			Entry("is 12hr format with AM postfix", "8:00 Am", "08:00"),
			Entry("is 12hr format with AM postfix", "08:00 aM", "08:00"),
			Entry("is 12hr format with PM postfix", "9:00 Pm", "21:00"),
			Entry("is 12hr format with PM postfix and extra padding", "09:00   pm", "21:00"),
			Entry("is 12hr format with minutes", "11:59   pM", "23:59"),
		)
	})

	Context("ParseAlertScheduleSettingsTimeHoursAndMinutes", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedHour int, expectedMinute int, expectedSuccess bool) {
				hour, minute, err := dexcom.ParseAlertScheduleSettingsTimeHoursAndMinutes(value)
				if expectedSuccess {
					Expect(err).ToNot(HaveOccurred())
				} else {
					Expect(err).To(MatchError("alert schedule settings time is not parsable"))
				}
				Expect(hour).To(Equal(expectedHour))
				Expect(minute).To(Equal(expectedMinute))
			},
			Entry("is an empty string", "", 0, 0, false),
			Entry("contains non-numbers", "a$: b", 0, 0, false),
			Entry("does not exactly match format", "1;23", 0, 0, false),
			Entry("has hour and minute minimum", "00:00", 0, 0, true),
			Entry("has hour and minute maximum", "99:99", 99, 99, true),
			Entry("is 12hr format with AM postfix", "8:00 Am", 8, 0, true),
			Entry("is 12hr format with AM postfix", "08:00 aM", 8, 0, true),
			Entry("is 12hr format with PM postfix", "9:00 Pm", 21, 0, true),
			Entry("is 12hr format with PM postfix and extra padding", "09:00   pm", 21, 0, true),
			Entry("is 12hr format with minutes", "11:59   pM", 23, 59, true),
		)
	})

	Context("IsValidAlertScheduleSettingsStartTime, AlertScheduleSettingsStartTimeValidator, and ValidateAlertScheduleSettingsStartTime", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(dexcom.IsValidAlertScheduleSettingsStartTime(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				dexcom.AlertScheduleSettingsStartTimeValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(dexcom.ValidateAlertScheduleSettingsStartTime(value), expectedErrors...)
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

	Context("IsValidAlertScheduleSettingsEndTime, AlertScheduleSettingsEndTimeValidator, and ValidateAlertScheduleSettingsEndTime", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(dexcom.IsValidAlertScheduleSettingsEndTime(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				dexcom.AlertScheduleSettingsEndTimeValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(dexcom.ValidateAlertScheduleSettingsEndTime(value), expectedErrors...)
			},
			Entry("is an empty string", "", structureValidator.ErrorValueEmpty()),
			Entry("contains non-numbers", "a$: b", dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid("a$: b")),
			Entry("does not exactly match format", "1;47", dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid("1;47")),
			Entry("has hour in range (lower)", "00:00"),
			Entry("has hour in range (upper)", "47:59"),
			Entry("has hour out of range (upper)", "48:00", dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid("48:00")),
			Entry("has minute in range (lower)", "00:00"),
			Entry("has minute in range (upper)", "47:59"),
			Entry("has minute out of range (upper)", "47:60", dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid("47:60")),
		)
	})

	Context("ErrorValueStringAsAlertScheduleSettingsTimeNotValid", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsAlertScheduleSettingsTimeNotValid with empty string", dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as alert schedule settings time`),
			Entry("is ErrorValueStringAsAlertScheduleSettingsTimeNotValid with non-empty string", dexcom.ErrorValueStringAsAlertScheduleSettingsTimeNotValid("XX:XX"), "value-not-valid", "value is not valid", `value "XX:XX" is not valid as alert schedule settings time`),
		)
	})
})
