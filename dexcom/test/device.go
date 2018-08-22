package test

import (
	"math/rand"
	"time"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomAlertSetting() *dexcom.AlertSetting {
	return RandomAlertSettingWithAlertName(test.RandomStringFromArray(dexcom.AlertNames()))
}

func RandomAlertSettingWithAlertName(alertName string) *dexcom.AlertSetting {
	datum := dexcom.NewAlertSetting()
	datum.SystemTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()))
	datum.DisplayTime = pointer.FromTime(test.RandomTime())
	datum.AlertName = pointer.FromString(alertName)
	switch *datum.AlertName {
	case dexcom.AlertNameFixedLow:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingFixedLowUnits()))
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromArray(dexcom.AlertSettingFixedLowValuesForUnits(datum.Unit)))
		datum.Delay = pointer.FromInt(0)
		datum.Snooze = pointer.FromInt(test.RandomIntFromArray(dexcom.AlertSettingFixedLowSnoozes()))
		datum.Enabled = pointer.FromBool(true)
	case dexcom.AlertNameLow:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingLowUnits()))
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromArray(dexcom.AlertSettingLowValuesForUnits(datum.Unit)))
		datum.Delay = pointer.FromInt(0)
		datum.Snooze = pointer.FromInt(test.RandomIntFromArray(dexcom.AlertSettingLowSnoozes()))
		datum.Enabled = pointer.FromBool(test.RandomBool())
	case dexcom.AlertNameHigh:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingHighUnits()))
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromArray(dexcom.AlertSettingHighValuesForUnits(datum.Unit)))
		datum.Delay = pointer.FromInt(0)
		datum.Snooze = pointer.FromInt(test.RandomIntFromArray(dexcom.AlertSettingHighSnoozes()))
		datum.Enabled = pointer.FromBool(test.RandomBool())
	case dexcom.AlertNameRise:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingRiseUnits()))
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromArray(dexcom.AlertSettingRiseValuesForUnits(datum.Unit)))
		datum.Delay = pointer.FromInt(0)
		datum.Snooze = pointer.FromInt(test.RandomIntFromArray(dexcom.AlertSettingRiseSnoozes()))
		datum.Enabled = pointer.FromBool(test.RandomBool())
	case dexcom.AlertNameFall:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingFallUnits()))
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromArray(dexcom.AlertSettingFallValuesForUnits(datum.Unit)))
		datum.Delay = pointer.FromInt(0)
		datum.Snooze = pointer.FromInt(test.RandomIntFromArray(dexcom.AlertSettingFallSnoozes()))
		datum.Enabled = pointer.FromBool(test.RandomBool())
	case dexcom.AlertNameOutOfRange:
		datum.Unit = pointer.FromString(test.RandomStringFromArray(dexcom.AlertSettingOutOfRangeUnits()))
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromArray(dexcom.AlertSettingOutOfRangeValuesForUnits(datum.Unit)))
		datum.Delay = pointer.FromInt(0)
		datum.Snooze = pointer.FromInt(test.RandomIntFromArray(dexcom.AlertSettingOutOfRangeSnoozes()))
		datum.Enabled = pointer.FromBool(test.RandomBool())
	}
	return datum
}

func CloneAlertSetting(datum *dexcom.AlertSetting) *dexcom.AlertSetting {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewAlertSetting()
	clone.SystemTime = pointer.CloneTime(datum.SystemTime)
	clone.DisplayTime = pointer.CloneTime(datum.DisplayTime)
	clone.AlertName = pointer.CloneString(datum.AlertName)
	clone.Unit = pointer.CloneString(datum.Unit)
	clone.Value = pointer.CloneFloat64(datum.Value)
	clone.Delay = pointer.CloneInt(datum.Delay)
	clone.Snooze = pointer.CloneInt(datum.Snooze)
	clone.Enabled = pointer.CloneBool(datum.Enabled)
	return clone
}

func RandomAlertSettings() *dexcom.AlertSettings {
	datum := dexcom.NewAlertSettings()
	for _, index := range rand.Perm(len(dexcom.AlertNames())) {
		*datum = append(*datum, RandomAlertSettingWithAlertName(dexcom.AlertNames()[index]))
	}
	return datum
}

func CloneAlertSettings(datum *dexcom.AlertSettings) *dexcom.AlertSettings {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewAlertSettings()
	for _, alertSetting := range *datum {
		*clone = append(*clone, CloneAlertSetting(alertSetting))
	}
	return clone
}
