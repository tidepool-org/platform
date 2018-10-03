package test

import (
	"math/rand"
	"time"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/test"
)

func RandomAlertSetting() *dexcom.AlertSetting {
	return RandomAlertSettingWithAlertName(test.RandomStringFromArray(dexcom.AlertNames()))
}

func RandomAlertSettingWithAlertName(alertName string) *dexcom.AlertSetting {
	datum := dexcom.NewAlertSetting()
	datum.SystemTime = test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now())
	datum.DisplayTime = test.RandomTime()
	datum.AlertName = alertName
	switch datum.AlertName {
	case dexcom.AlertNameFixedLow:
		datum.Unit = test.RandomStringFromArray(dexcom.AlertSettingFixedLowUnits())
		datum.Value = test.RandomFloat64FromArray(dexcom.AlertSettingFixedLowValuesForUnits(datum.Unit))
		datum.Delay = 0
		datum.Snooze = test.RandomIntFromArray(dexcom.AlertSettingFixedLowSnoozes())
		datum.Enabled = true
	case dexcom.AlertNameLow:
		datum.Unit = test.RandomStringFromArray(dexcom.AlertSettingLowUnits())
		datum.Value = test.RandomFloat64FromArray(dexcom.AlertSettingLowValuesForUnits(datum.Unit))
		datum.Delay = 0
		datum.Snooze = test.RandomIntFromArray(dexcom.AlertSettingLowSnoozes())
		datum.Enabled = test.RandomBool()
	case dexcom.AlertNameHigh:
		datum.Unit = test.RandomStringFromArray(dexcom.AlertSettingHighUnits())
		datum.Value = test.RandomFloat64FromArray(dexcom.AlertSettingHighValuesForUnits(datum.Unit))
		datum.Delay = 0
		datum.Snooze = test.RandomIntFromArray(dexcom.AlertSettingHighSnoozes())
		datum.Enabled = test.RandomBool()
	case dexcom.AlertNameRise:
		datum.Unit = test.RandomStringFromArray(dexcom.AlertSettingRiseUnits())
		datum.Value = test.RandomFloat64FromArray(dexcom.AlertSettingRiseValuesForUnits(datum.Unit))
		datum.Delay = 0
		datum.Snooze = test.RandomIntFromArray(dexcom.AlertSettingRiseSnoozes())
		datum.Enabled = test.RandomBool()
	case dexcom.AlertNameFall:
		datum.Unit = test.RandomStringFromArray(dexcom.AlertSettingFallUnits())
		datum.Value = test.RandomFloat64FromArray(dexcom.AlertSettingFallValuesForUnits(datum.Unit))
		datum.Delay = 0
		datum.Snooze = test.RandomIntFromArray(dexcom.AlertSettingFallSnoozes())
		datum.Enabled = test.RandomBool()
	case dexcom.AlertNameOutOfRange:
		datum.Unit = test.RandomStringFromArray(dexcom.AlertSettingOutOfRangeUnits())
		datum.Value = test.RandomFloat64FromArray(dexcom.AlertSettingOutOfRangeValuesForUnits(datum.Unit))
		datum.Delay = 0
		datum.Snooze = test.RandomIntFromArray(dexcom.AlertSettingOutOfRangeSnoozes())
		datum.Enabled = test.RandomBool()
	}
	return datum
}

func CloneAlertSetting(datum *dexcom.AlertSetting) *dexcom.AlertSetting {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewAlertSetting()
	clone.SystemTime = datum.SystemTime
	clone.DisplayTime = datum.DisplayTime
	clone.AlertName = datum.AlertName
	clone.Unit = datum.Unit
	clone.Value = datum.Value
	clone.Delay = datum.Delay
	clone.Snooze = datum.Snooze
	clone.Enabled = datum.Enabled
	return clone
}

func RandomAlertSettings() dexcom.AlertSettings {
	datum := dexcom.AlertSettings{}
	for _, index := range rand.Perm(len(dexcom.AlertNames())) {
		datum = append(datum, RandomAlertSettingWithAlertName(dexcom.AlertNames()[index]))
	}
	return datum
}

func CloneAlertSettings(datum dexcom.AlertSettings) dexcom.AlertSettings {
	if datum == nil {
		return nil
	}
	clone := dexcom.AlertSettings{}
	for _, alertSetting := range datum {
		clone = append(clone, CloneAlertSetting(alertSetting))
	}
	return clone
}
