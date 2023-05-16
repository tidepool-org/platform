package test

import (
	"fmt"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomAlertSchedules(minimumLength int, maximumLength int) *dexcom.AlertSchedules {
	datum := make(dexcom.AlertSchedules, test.RandomIntFromRange(minimumLength, maximumLength))
	if length := len(datum); length > 0 {
		defaultIndex := test.RandomIntFromRange(0, length-1)
		for index := range datum {
			datum[index] = RandomAlertSchedule(index == defaultIndex)
		}
	}
	return &datum
}

func CloneAlertSchedules(datum *dexcom.AlertSchedules) *dexcom.AlertSchedules {
	if datum == nil {
		return nil
	}
	clone := make(dexcom.AlertSchedules, len(*datum))
	for index, d := range *datum {
		clone[index] = CloneAlertSchedule(d)
	}
	return &clone
}

func RandomAlertSchedule(isDefault bool) *dexcom.AlertSchedule {
	datum := dexcom.NewAlertSchedule()
	datum.AlertScheduleSettings = RandomAlertScheduleSettings(isDefault)
	datum.AlertSettings = RandomAlertSettings(1, 3)
	return datum
}

func CloneAlertSchedule(datum *dexcom.AlertSchedule) *dexcom.AlertSchedule {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewAlertSchedule()
	clone.AlertScheduleSettings = CloneAlertScheduleSettings(datum.AlertScheduleSettings)
	clone.AlertSettings = CloneAlertSettings(datum.AlertSettings)
	return clone
}

func RandomAlertScheduleSettings(isDefault bool) *dexcom.AlertScheduleSettings {
	datum := dexcom.NewAlertScheduleSettings()
	if isDefault {
		datum.Name = pointer.FromString("")
		datum.Enabled = pointer.FromBool(test.RandomBool())
		datum.Default = pointer.FromBool(true)
		datum.StartTime = pointer.FromString(dexcom.AlertScheduleSettingsStartTimeDefault)
		datum.EndTime = pointer.FromString(dexcom.AlertScheduleSettingsEndTimeDefault)
		datum.DaysOfWeek = pointer.FromStringArray(dexcom.AlertScheduleSettingsDays())
	} else {
		datum.Name = pointer.FromString(RandomAlertScheduleSettingsName())
		datum.Enabled = pointer.FromBool(test.RandomBool())
		datum.Default = pointer.FromBool(false)
		datum.StartTime = pointer.FromString(fmt.Sprintf("%02d:%02d", test.RandomIntFromRange(0, 23), test.RandomIntFromRange(0, 59)))
		datum.EndTime = pointer.FromString(fmt.Sprintf("%02d:%02d", test.RandomIntFromRange(0, 23), test.RandomIntFromRange(0, 59)))
		datum.DaysOfWeek = pointer.FromStringArray(test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(0, len(dexcom.AlertScheduleSettingsDays()), dexcom.AlertScheduleSettingsDays()))
	}
	return datum
}

func CloneAlertScheduleSettings(datum *dexcom.AlertScheduleSettings) *dexcom.AlertScheduleSettings {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewAlertScheduleSettings()
	clone.Name = pointer.CloneString(datum.Name)
	clone.Enabled = pointer.CloneBool(datum.Enabled)
	clone.Default = pointer.CloneBool(datum.Default)
	clone.StartTime = pointer.CloneString(datum.StartTime)
	clone.EndTime = pointer.CloneString(datum.EndTime)
	clone.DaysOfWeek = pointer.CloneStringArray(datum.DaysOfWeek)
	return clone
}

func RandomAlertSettings(minimumLength int, maximumLength int) *dexcom.AlertSettings {
	datum := make(dexcom.AlertSettings, test.RandomIntFromRange(minimumLength, maximumLength))
	for index := range datum {
		datum[index] = RandomAlertSetting(nil)
	}
	datum.Deduplicate()
	return &datum
}

func CloneAlertSettings(datum *dexcom.AlertSettings) *dexcom.AlertSettings {
	if datum == nil {
		return nil
	}
	clone := make(dexcom.AlertSettings, len(*datum))
	for index, d := range *datum {
		clone[index] = CloneAlertSetting(d)
	}
	return &clone
}

func RandomAlertSetting(withName *string) *dexcom.AlertSetting {
	datum := dexcom.NewAlertSetting()
	datum.SystemTime = RandomSystemTime()
	datum.DisplayTime = RandomDisplayTime()
	if withName != nil {
		datum.AlertName = withName
	} else {
		datum.AlertName = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingAlertNames()))
	}
	switch *datum.AlertName {
	case dexcom.AlertSettingAlertNameFall:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingUnitFalls()))
		switch *datum.Unit {
		case dexcom.AlertSettingUnitMgdLMinute:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueFallMgdLMinuteMinimum, dexcom.AlertSettingValueFallMgdLMinuteMaximum))
		}
		datum.Snooze = pointer.FromInt(test.RandomIntFromRange(dexcom.AlertSettingSnoozeMinutesMinimum, dexcom.AlertSettingSnoozeMinutesMaximum))
		datum.Enabled = pointer.FromBool(test.RandomBool())
	case dexcom.AlertSettingAlertNameHigh:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingUnitHighs()))
		switch *datum.Unit {
		case dexcom.AlertSettingUnitMgdL:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueHighMgdLMinimum, dexcom.AlertSettingValueHighMgdLMaximum))
		}
		datum.Snooze = pointer.FromInt(test.RandomIntFromRange(dexcom.AlertSettingSnoozeMinutesMinimum, dexcom.AlertSettingSnoozeMinutesMaximum))
		datum.Enabled = pointer.FromBool(test.RandomBool())
	case dexcom.AlertSettingAlertNameLow:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingUnitLows()))
		switch *datum.Unit {
		case dexcom.AlertSettingUnitMgdL:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueLowMgdLMinimum, dexcom.AlertSettingValueLowMgdLMaximum))
		}
		datum.Snooze = pointer.FromInt(test.RandomIntFromRange(dexcom.AlertSettingSnoozeMinutesMinimum, dexcom.AlertSettingSnoozeMinutesMaximum))
		datum.Enabled = pointer.FromBool(test.RandomBool())
	case dexcom.AlertSettingAlertNameNoReadings:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingUnitNoReadings()))
		switch *datum.Unit {
		case dexcom.AlertSettingUnitMinutes:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueNoReadingsMgdLMinimum, dexcom.AlertSettingValueNoReadingsMgdLMaximum))
		}
		datum.Snooze = pointer.FromInt(test.RandomIntFromRange(dexcom.AlertSettingSnoozeMinutesMinimum, dexcom.AlertSettingSnoozeMinutesMaximum))
		datum.Enabled = pointer.FromBool(test.RandomBool())
	case dexcom.AlertSettingAlertNameOutOfRange:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingUnitOutOfRanges()))
		switch *datum.Unit {
		case dexcom.AlertSettingUnitMinutes:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueOutOfRangeMgdLMinimum, dexcom.AlertSettingValueOutOfRangeMgdLMaximum))
		}
		datum.Snooze = pointer.FromInt(test.RandomIntFromRange(dexcom.AlertSettingSnoozeMinutesMinimum, dexcom.AlertSettingSnoozeMinutesMaximum))
		datum.Enabled = pointer.FromBool(test.RandomBool())
	case dexcom.AlertSettingAlertNameRise:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingUnitRises()))
		switch *datum.Unit {
		case dexcom.AlertSettingUnitMgdLMinute:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueRiseMgdLMinuteMinimum, dexcom.AlertSettingValueRiseMgdLMinuteMaximum))
		}
		datum.Snooze = pointer.FromInt(test.RandomIntFromRange(dexcom.AlertSettingSnoozeMinutesMinimum, dexcom.AlertSettingSnoozeMinutesMaximum))
		datum.Enabled = pointer.FromBool(test.RandomBool())
	case dexcom.AlertSettingAlertNameUrgentLow:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingUnitUrgentLows()))
		switch *datum.Unit {
		case dexcom.AlertSettingUnitMgdL:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueUrgentLowMgdLMinimum, dexcom.AlertSettingValueUrgentLowMgdLMaximum))
		}
		datum.Snooze = pointer.FromInt(test.RandomIntFromRange(dexcom.AlertSettingSnoozeMinutesMinimum, dexcom.AlertSettingSnoozeMinutesMaximum))
		datum.Enabled = pointer.FromBool(true)
	case dexcom.AlertSettingAlertNameUrgentLowSoon:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingUnitUrgentLowSoons()))
		switch *datum.Unit {
		case dexcom.AlertSettingUnitMgdL:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueUrgentLowSoonMgdLMinimum, dexcom.AlertSettingValueUrgentLowSoonMgdLMaximum))
		}
		datum.Snooze = pointer.FromInt(test.RandomIntFromRange(dexcom.AlertSettingSnoozeMinutesMinimum, dexcom.AlertSettingSnoozeMinutesMaximum))
		datum.Enabled = pointer.FromBool(test.RandomBool())
	case dexcom.AlertSettingAlertNameFixedLow:
		datum.Enabled = pointer.FromBool(test.RandomBool())
	case dexcom.AlertSettingAlertNameUnknown:
		datum.Unit = pointer.FromString(dexcom.AlertSettingUnitUnknown)
		datum.Enabled = pointer.FromBool(test.RandomBool())
	}
	return datum
}

func CloneAlertSetting(datum *dexcom.AlertSetting) *dexcom.AlertSetting {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewAlertSetting()
	clone.SystemTime = CloneTime(datum.SystemTime)
	clone.DisplayTime = CloneTime(datum.DisplayTime)
	clone.AlertName = pointer.CloneString(datum.AlertName)
	clone.Unit = pointer.CloneString(datum.Unit)
	clone.Value = pointer.CloneFloat64(datum.Value)
	clone.Snooze = pointer.CloneInt(datum.Snooze)
	clone.Enabled = pointer.CloneBool(datum.Enabled)
	return clone
}

func RandomAlertScheduleSettingsName() string {
	return test.RandomString()
}
