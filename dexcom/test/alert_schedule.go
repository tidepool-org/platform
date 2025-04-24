package test

import (
	"fmt"
	"math"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomAlertSchedules(minimumLength int, maximumLength int) *dexcom.AlertSchedules {
	datum := make(dexcom.AlertSchedules, test.RandomIntFromRange(minimumLength, maximumLength))
	if length := len(datum); length > 0 {
		defaultIndex := test.RandomIntFromRange(0, length-1)
		for index := range datum {
			datum[index] = RandomAlertScheduleWithDefault(index == defaultIndex)
		}
	}
	return &datum
}

func CloneAlertSchedules(datum *dexcom.AlertSchedules) *dexcom.AlertSchedules {
	if datum == nil {
		return nil
	}
	clone := make(dexcom.AlertSchedules, len(*datum))
	for index, datum := range *datum {
		clone[index] = CloneAlertSchedule(datum)
	}
	return &clone
}

func NewArrayFromAlertSchedules(datumArray *dexcom.AlertSchedules, objectFormat test.ObjectFormat) []interface{} {
	if datumArray == nil {
		return nil
	}
	array := make([]interface{}, len(*datumArray))
	for index, datum := range *datumArray {
		array[index] = NewObjectFromAlertSchedule(datum, objectFormat)
	}
	return array
}

func RandomAlertSchedule() *dexcom.AlertSchedule {
	return RandomAlertScheduleWithDefault(test.RandomBool())
}

func RandomAlertScheduleWithDefault(isDefault bool) *dexcom.AlertSchedule {
	datum := dexcom.NewAlertSchedule()
	datum.AlertScheduleSettings = RandomAlertScheduleSettingsWithDefault(isDefault)
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

func NewObjectFromAlertSchedule(datum *dexcom.AlertSchedule, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.AlertScheduleSettings != nil {
		object["alertScheduleSettings"] = NewObjectFromAlertScheduleSettings(datum.AlertScheduleSettings, objectFormat)
	}
	if datum.AlertSettings != nil {
		object["alertSettings"] = NewArrayFromAlertSettings(datum.AlertSettings, objectFormat)
	}
	return object
}

func RandomAlertScheduleSettings() *dexcom.AlertScheduleSettings {
	return RandomAlertScheduleSettingsWithDefault(test.RandomBool())
}

func RandomAlertScheduleSettingsWithDefault(isDefault bool) *dexcom.AlertScheduleSettings {
	datum := dexcom.NewAlertScheduleSettings()
	datum.IsDefaultSchedule = pointer.FromBool(isDefault)
	datum.IsEnabled = pointer.FromBool(test.RandomBool())
	datum.IsActive = pointer.FromBool(test.RandomBool())
	if isDefault {
		datum.AlertScheduleName = pointer.FromString("")
		datum.StartTime = pointer.FromString(dexcom.AlertScheduleSettingsStartTimeDefault)
		datum.EndTime = pointer.FromString(dexcom.AlertScheduleSettingsEndTimeDefault)
		datum.DaysOfWeek = pointer.FromStringArray(dexcom.AlertScheduleSettingsDays())
	} else {
		datum.AlertScheduleName = pointer.FromString(test.RandomString())
		datum.StartTime = pointer.FromString(fmt.Sprintf("%02d:%02d", test.RandomIntFromRange(0, 23), test.RandomIntFromRange(0, 59)))
		datum.EndTime = pointer.FromString(fmt.Sprintf("%02d:%02d", test.RandomIntFromRange(0, 23), test.RandomIntFromRange(0, 59)))
		datum.DaysOfWeek = pointer.FromStringArray(test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(0, len(dexcom.AlertScheduleSettingsDays()), dexcom.AlertScheduleSettingsDays()))
		datum.Override = RandomOverride()
	}
	return datum
}

func CloneAlertScheduleSettings(datum *dexcom.AlertScheduleSettings) *dexcom.AlertScheduleSettings {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewAlertScheduleSettings()
	clone.IsDefaultSchedule = pointer.CloneBool(datum.IsDefaultSchedule)
	clone.IsEnabled = pointer.CloneBool(datum.IsEnabled)
	clone.IsActive = pointer.CloneBool(datum.IsActive)
	clone.AlertScheduleName = pointer.CloneString(datum.AlertScheduleName)
	clone.StartTime = pointer.CloneString(datum.StartTime)
	clone.EndTime = pointer.CloneString(datum.EndTime)
	clone.DaysOfWeek = pointer.CloneStringArray(datum.DaysOfWeek)
	clone.Override = CloneOverride(datum.Override)
	return clone
}

func NewObjectFromAlertScheduleSettings(datum *dexcom.AlertScheduleSettings, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.IsDefaultSchedule != nil {
		object["isDefaultSchedule"] = test.NewObjectFromBool(*datum.IsDefaultSchedule, objectFormat)
	}
	if datum.IsEnabled != nil {
		object["isEnabled"] = test.NewObjectFromBool(*datum.IsEnabled, objectFormat)
	}
	if datum.IsActive != nil {
		object["isActive"] = test.NewObjectFromBool(*datum.IsActive, objectFormat)
	}
	if datum.AlertScheduleName != nil {
		object["alertScheduleName"] = test.NewObjectFromString(*datum.AlertScheduleName, objectFormat)
	}
	if datum.StartTime != nil {
		object["startTime"] = test.NewObjectFromString(*datum.StartTime, objectFormat)
	}
	if datum.EndTime != nil {
		object["endTime"] = test.NewObjectFromString(*datum.EndTime, objectFormat)
	}
	if datum.DaysOfWeek != nil {
		object["daysOfWeek"] = test.NewObjectFromStringArray(*datum.DaysOfWeek, objectFormat)
	}
	if datum.Override != nil {
		object["override"] = NewObjectFromOverride(datum.Override, objectFormat)
	}
	return object
}

func RandomOverride() *dexcom.Override {
	datum := dexcom.NewOverride()
	datum.IsOverrideEnabled = pointer.FromBool(test.RandomBool())
	datum.Mode = pointer.FromString(test.RandomStringFromArray(dexcom.AlertScheduleSettingsOverrideModes()))
	datum.EndTime = RandomTime()
	return datum
}

func CloneOverride(datum *dexcom.Override) *dexcom.Override {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewOverride()
	clone.IsOverrideEnabled = pointer.CloneBool(datum.IsOverrideEnabled)
	clone.Mode = pointer.CloneString(datum.Mode)
	clone.EndTime = CloneTime(datum.EndTime)
	return clone
}

func NewObjectFromOverride(datum *dexcom.Override, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.IsOverrideEnabled != nil {
		object["isOverrideEnabled"] = test.NewObjectFromBool(*datum.IsOverrideEnabled, objectFormat)
	}
	if datum.Mode != nil {
		object["mode"] = test.NewObjectFromString(*datum.Mode, objectFormat)
	}
	if datum.EndTime != nil {
		object["endTime"] = test.NewObjectFromString(datum.EndTime.String(), objectFormat)
	}
	return object
}

func RandomAlertSettings(minimumLength int, maximumLength int) *dexcom.AlertSettings {
	datum := make(dexcom.AlertSettings, test.RandomIntFromRange(minimumLength, maximumLength))
	for index := range datum {
		datum[index] = RandomAlertSetting()
	}
	datum.Deduplicate()
	return &datum
}

func CloneAlertSettings(datum *dexcom.AlertSettings) *dexcom.AlertSettings {
	if datum == nil {
		return nil
	}
	clone := make(dexcom.AlertSettings, len(*datum))
	for index, datum := range *datum {
		clone[index] = CloneAlertSetting(datum)
	}
	return &clone
}

func NewArrayFromAlertSettings(datumArray *dexcom.AlertSettings, objectFormat test.ObjectFormat) []interface{} {
	if datumArray == nil {
		return nil
	}
	array := make([]interface{}, len(*datumArray))
	for index, datum := range *datumArray {
		array[index] = NewObjectFromAlertSetting(datum, objectFormat)
	}
	return array
}

func RandomAlertSetting() *dexcom.AlertSetting {
	datum := dexcom.NewAlertSetting()
	datum.SystemTime = RandomSystemTime()
	datum.DisplayTime = RandomDisplayTime()
	datum.AlertName = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingAlertNames()))
	switch *datum.AlertName {
	case dexcom.AlertSettingAlertNameUnknown:
		datum.Unit = pointer.FromString(dexcom.AlertSettingUnitUnknown)
		datum.Enabled = pointer.FromBool(test.RandomBool())
	case dexcom.AlertSettingAlertNameHigh:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingUnitHighs()))
		switch *datum.Unit {
		case dexcom.AlertSettingUnitMgdL:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueHighMgdLMinimum, dexcom.AlertSettingValueHighMgdLMaximum))
		case dexcom.AlertSettingUnitMmolL:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueHighMmolLMinimum, dexcom.AlertSettingValueHighMmolLMaximum))
		}
		datum.Snooze = pointer.FromInt(test.RandomIntFromRange(dexcom.AlertSettingSnoozeMinutesMinimum, dexcom.AlertSettingSnoozeMinutesMaximum))
		datum.Enabled = pointer.FromBool(test.RandomBool())
		datum.Delay = pointer.FromInt(test.RandomIntFromRange(dexcom.AlertSettingDelayMinimum, math.MaxInt32))
		datum.SecondaryTriggerCondition = pointer.FromInt(test.RandomInt())
	case dexcom.AlertSettingAlertNameLow:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingUnitLows()))
		switch *datum.Unit {
		case dexcom.AlertSettingUnitMgdL:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueLowMgdLMinimum, dexcom.AlertSettingValueLowMgdLMaximum))
		case dexcom.AlertSettingUnitMmolL:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueLowMmolLMinimum, dexcom.AlertSettingValueLowMmolLMaximum))
		}
		datum.Snooze = pointer.FromInt(test.RandomIntFromRange(dexcom.AlertSettingSnoozeMinutesMinimum, dexcom.AlertSettingSnoozeMinutesMaximum))
		datum.Enabled = pointer.FromBool(test.RandomBool())
		datum.SecondaryTriggerCondition = pointer.FromInt(test.RandomInt())
	case dexcom.AlertSettingAlertNameRise:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingUnitRises()))
		switch *datum.Unit {
		case dexcom.AlertSettingUnitMgdLMinute:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueRiseMgdLMinuteMinimum, dexcom.AlertSettingValueRiseMgdLMinuteMaximum))
		case dexcom.AlertSettingUnitMmolLMinute:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueRiseMmolLMinuteMinimum, dexcom.AlertSettingValueRiseMmolLMinuteMaximum))
		}
		datum.Enabled = pointer.FromBool(test.RandomBool())
		datum.SecondaryTriggerCondition = pointer.FromInt(test.RandomInt())
	case dexcom.AlertSettingAlertNameFall:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingUnitFalls()))
		switch *datum.Unit {
		case dexcom.AlertSettingUnitMgdLMinute:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueFallMgdLMinuteMinimum, dexcom.AlertSettingValueFallMgdLMinuteMaximum))
		case dexcom.AlertSettingUnitMmolLMinute:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueFallMmolLMinuteMinimum, dexcom.AlertSettingValueFallMmolLMinuteMaximum))
		}
		datum.Snooze = pointer.FromInt(test.RandomIntFromRange(dexcom.AlertSettingSnoozeMinutesMinimum, dexcom.AlertSettingSnoozeMinutesMaximum))
		datum.Enabled = pointer.FromBool(test.RandomBool())
		datum.SecondaryTriggerCondition = pointer.FromInt(test.RandomInt())
	case dexcom.AlertSettingAlertNameOutOfRange:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingUnitOutOfRanges()))
		switch *datum.Unit {
		case dexcom.AlertSettingUnitMinutes:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueOutOfRangeMinutesMinimum, dexcom.AlertSettingValueOutOfRangeMinutesMaximum))
		}
		datum.Enabled = pointer.FromBool(test.RandomBool())
		datum.SecondaryTriggerCondition = pointer.FromInt(test.RandomInt())
	case dexcom.AlertSettingAlertNameUrgentLow:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingUnitUrgentLows()))
		switch *datum.Unit {
		case dexcom.AlertSettingUnitMgdL:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueUrgentLowMgdLMinimum, dexcom.AlertSettingValueUrgentLowMgdLMaximum))
		case dexcom.AlertSettingUnitMmolL:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueUrgentLowMmolLMinimum, dexcom.AlertSettingValueUrgentLowMmolLMaximum))
		}
		datum.Snooze = pointer.FromInt(test.RandomIntFromRange(dexcom.AlertSettingSnoozeMinutesMinimum, dexcom.AlertSettingSnoozeMinutesMaximum))
		datum.Enabled = pointer.FromBool(true)
		datum.SecondaryTriggerCondition = pointer.FromInt(test.RandomInt())
	case dexcom.AlertSettingAlertNameUrgentLowSoon:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingUnitUrgentLowSoons()))
		switch *datum.Unit {
		case dexcom.AlertSettingUnitMgdL:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueUrgentLowSoonMgdLMinimum, dexcom.AlertSettingValueUrgentLowSoonMgdLMaximum))
		case dexcom.AlertSettingUnitMmolL:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueUrgentLowSoonMmolLMinimum, dexcom.AlertSettingValueUrgentLowSoonMmolLMaximum))
		}
		datum.Snooze = pointer.FromInt(test.RandomIntFromRange(dexcom.AlertSettingSnoozeMinutesMinimum, dexcom.AlertSettingSnoozeMinutesMaximum))
		datum.Enabled = pointer.FromBool(test.RandomBool())
		datum.SecondaryTriggerCondition = pointer.FromInt(test.RandomInt())
	case dexcom.AlertSettingAlertNameNoReadings:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingUnitNoReadings()))
		switch *datum.Unit {
		case dexcom.AlertSettingUnitMinutes:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueNoReadingsMinutesMinimum, dexcom.AlertSettingValueNoReadingsMinutesMaximum))
		}
		datum.Snooze = pointer.FromInt(test.RandomIntFromRange(dexcom.AlertSettingSnoozeMinutesMinimum, dexcom.AlertSettingSnoozeMinutesMaximum))
		datum.Enabled = pointer.FromBool(test.RandomBool())
		datum.SecondaryTriggerCondition = pointer.FromInt(test.RandomInt())
	case dexcom.AlertSettingAlertNameFixedLow:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingUnitFixedLows()))
		switch *datum.Unit {
		case dexcom.AlertSettingUnitMinutes:
			datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dexcom.AlertSettingValueNoReadingsMinutesMinimum, dexcom.AlertSettingValueNoReadingsMinutesMaximum))
		}
		datum.Enabled = pointer.FromBool(true)
	}
	datum.SoundTheme = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingSoundThemes()))
	datum.SoundOutputMode = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingSoundOutputModes()))
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
	clone.Delay = pointer.CloneInt(datum.Delay)
	clone.SecondaryTriggerCondition = pointer.CloneInt(datum.SecondaryTriggerCondition)
	clone.SoundTheme = pointer.CloneString(datum.SoundTheme)
	clone.SoundOutputMode = pointer.CloneString(datum.SoundOutputMode)
	return clone
}

func NewObjectFromAlertSetting(datum *dexcom.AlertSetting, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.SystemTime != nil {
		object["systemTime"] = test.NewObjectFromString(datum.SystemTime.String(), objectFormat)
	}
	if datum.DisplayTime != nil {
		object["displayTime"] = test.NewObjectFromString(datum.DisplayTime.String(), objectFormat)
	}
	if datum.AlertName != nil {
		object["alertName"] = test.NewObjectFromString(*datum.AlertName, objectFormat)
	}
	if datum.Unit != nil {
		object["unit"] = test.NewObjectFromString(*datum.Unit, objectFormat)
	}
	if datum.Value != nil {
		object["value"] = test.NewObjectFromFloat64(*datum.Value, objectFormat)
	}
	if datum.Snooze != nil {
		object["snooze"] = test.NewObjectFromInt(*datum.Snooze, objectFormat)
	}
	if datum.Enabled != nil {
		object["enabled"] = test.NewObjectFromBool(*datum.Enabled, objectFormat)
	}
	if datum.Delay != nil {
		object["delay"] = test.NewObjectFromInt(*datum.Delay, objectFormat)
	}
	if datum.SecondaryTriggerCondition != nil {
		object["secondaryTriggerCondition"] = test.NewObjectFromInt(*datum.SecondaryTriggerCondition, objectFormat)
	}
	if datum.SoundTheme != nil {
		object["soundTheme"] = test.NewObjectFromString(*datum.SoundTheme, objectFormat)
	}
	if datum.SoundOutputMode != nil {
		object["soundOutputMode"] = test.NewObjectFromString(*datum.SoundOutputMode, objectFormat)
	}
	return object
}
