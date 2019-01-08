package fetch

import (
	"time"

	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesActivityPhysical "github.com/tidepool-org/platform/data/types/activity/physical"
	dataTypesBloodGlucoseContinuous "github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	dataTypesCommonOrigin "github.com/tidepool-org/platform/data/types/common/origin"
	dataTypesDeviceCalibration "github.com/tidepool-org/platform/data/types/device/calibration"
	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	dataTypesInsulin "github.com/tidepool-org/platform/data/types/insulin"
	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	dataTypesStateReported "github.com/tidepool-org/platform/data/types/state/reported"
	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/pointer"
)

// TODO: For now this assumes that the systemTime is close to true UTC time (+/- some small drift).
// However, it is possible for this to NOT be true if the device receives a hard reset.
// Unfortunately, the only way to detect that MIGHT be to look between multiple events.
// If there is a large gap between systemTimes, and a much larger or smaller gap between displayTimes,
// then it MIGHT indicate a hard reset. (It may also simply represent and period of time where the
// device was not in use and displayTime immediately prior to or immediately after period not it use
// were grossly in error.)

const OffsetDuration = 30 * time.Minute // Duration between time zone offsets we scan for

const MaximumOffsets = (14 * time.Hour) / OffsetDuration  // Maximum time zone offset is +14:00
const MinimumOffsets = (-12 * time.Hour) / OffsetDuration // Minimum time zone offset is -12:00

const DailyDuration = 24 * time.Hour
const DailyOffsets = DailyDuration / OffsetDuration

func translateTime(systemTime *dexcom.Time, displayTime *dexcom.Time, datum *dataTypes.Base) {
	var clockDriftOffsetDuration time.Duration
	var conversionOffsetDuration time.Duration
	var timeZoneOffsetDuration time.Duration

	delta := displayTime.Sub(*systemTime.Raw())
	if delta > 0 {
		offsetCount := time.Duration((float64(delta) + float64(OffsetDuration)/2) / float64(OffsetDuration))
		clockDriftOffsetDuration = delta - offsetCount*OffsetDuration
		for offsetCount > MaximumOffsets {
			conversionOffsetDuration += DailyDuration
			offsetCount -= DailyOffsets
		}
		timeZoneOffsetDuration = offsetCount * OffsetDuration
	} else if delta < 0 {
		offsetCount := time.Duration((float64(delta) - float64(OffsetDuration)/2) / float64(OffsetDuration))
		clockDriftOffsetDuration = delta - offsetCount*OffsetDuration
		for offsetCount < MinimumOffsets {
			conversionOffsetDuration -= DailyDuration
			offsetCount += DailyOffsets
		}
		timeZoneOffsetDuration = offsetCount * OffsetDuration
	}

	datum.Time = pointer.FromString(systemTime.Format(dataTypes.TimeFormat))
	datum.DeviceTime = pointer.FromString(displayTime.Format(dataTypes.DeviceTimeFormat))
	datum.TimeZoneOffset = pointer.FromInt(int(timeZoneOffsetDuration / time.Minute))
	if clockDriftOffsetDuration != 0 {
		datum.ClockDriftOffset = pointer.FromInt(int(clockDriftOffsetDuration / time.Millisecond))
	}
	if conversionOffsetDuration != 0 {
		datum.ConversionOffset = pointer.FromInt(int(conversionOffsetDuration / time.Millisecond))
	}

	if datum.Payload == nil {
		datum.Payload = data.NewBlob()
	}
	(*datum.Payload)["systemTime"] = systemTime.Raw()
}

func translateCalibrationToDatum(calibration *dexcom.Calibration) data.Datum {
	datum := dataTypesDeviceCalibration.New()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = nil
	datum.GUID = nil

	datum.Value = pointer.CloneFloat64(calibration.Value)
	datum.Units = pointer.CloneString(calibration.Unit)
	datum.Payload = data.NewBlob()
	if calibration.TransmitterID != nil {
		(*datum.Payload)["transmitterId"] = *calibration.TransmitterID
	}

	translateTime(calibration.SystemTime, calibration.DisplayTime, &datum.Base)
	return datum
}

func translateDeviceToDatum(device *dexcom.Device) data.Datum {
	datum := dataTypesSettingsCgm.New()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = nil
	datum.GUID = nil

	datum.Manufacturers = pointer.FromStringArray([]string{"Dexcom"})
	datum.SerialNumber = pointer.CloneString(device.SerialNumber)
	datum.TransmitterID = pointer.CloneString(device.TransmitterID)
	datum.Units = pointer.FromString(dataBloodGlucose.MgdL)

	defaultAlertSchedule := device.AlertScheduleList.Default()
	if defaultAlertSchedule != nil {
		datum.DefaultAlerts = translateAlertSettingsToAlerts(defaultAlertSchedule.AlertScheduleSettings.Enabled, defaultAlertSchedule.AlertSettings)
		for _, alertSetting := range *defaultAlertSchedule.AlertSettings {
			switch *alertSetting.AlertName {
			case dexcom.AlertSettingAlertNameFall:
				if datum.RateAlerts == nil {
					datum.RateAlerts = dataTypesSettingsCgm.NewRateAlertsDEPRECATED()
				}
				datum.RateAlerts.FallRateAlert = dataTypesSettingsCgm.NewFallRateAlertDEPRECATED()
				datum.RateAlerts.FallRateAlert.Enabled = pointer.CloneBool(alertSetting.Enabled)
				datum.RateAlerts.FallRateAlert.Rate = pointer.FromFloat64(-*alertSetting.Value)
			case dexcom.AlertSettingAlertNameHigh:
				datum.HighLevelAlert = dataTypesSettingsCgm.NewHighLevelAlertDEPRECATED()
				datum.HighLevelAlert.Enabled = pointer.CloneBool(alertSetting.Enabled)
				datum.HighLevelAlert.Level = pointer.CloneFloat64(alertSetting.Value)
				datum.HighLevelAlert.Snooze = pointer.FromInt(*alertSetting.Snooze * 60 * 1000)
			case dexcom.AlertSettingAlertNameLow:
				datum.LowLevelAlert = dataTypesSettingsCgm.NewLowLevelAlertDEPRECATED()
				datum.LowLevelAlert.Enabled = pointer.CloneBool(alertSetting.Enabled)
				datum.LowLevelAlert.Level = pointer.CloneFloat64(alertSetting.Value)
				datum.LowLevelAlert.Snooze = pointer.FromInt(*alertSetting.Snooze * 60 * 1000)
			case dexcom.AlertSettingAlertNameOutOfRange:
				datum.OutOfRangeAlert = dataTypesSettingsCgm.NewOutOfRangeAlertDEPRECATED()
				datum.OutOfRangeAlert.Enabled = pointer.CloneBool(alertSetting.Enabled)
				datum.OutOfRangeAlert.Threshold = pointer.FromInt(int(*alertSetting.Value) * 60 * 1000)
			case dexcom.AlertSettingAlertNameRise:
				if datum.RateAlerts == nil {
					datum.RateAlerts = dataTypesSettingsCgm.NewRateAlertsDEPRECATED()
				}
				datum.RateAlerts.RiseRateAlert = dataTypesSettingsCgm.NewRiseRateAlertDEPRECATED()
				datum.RateAlerts.RiseRateAlert.Enabled = pointer.CloneBool(alertSetting.Enabled)
				datum.RateAlerts.RiseRateAlert.Rate = pointer.CloneFloat64(alertSetting.Value)
			}
		}
	}

	var scheduledAlerts dataTypesSettingsCgm.ScheduledAlerts
	for _, alertSchedule := range *device.AlertScheduleList {
		if alertSchedule != defaultAlertSchedule {
			scheduledAlerts = append(scheduledAlerts, translateAlertScheduleToScheduledAlert(alertSchedule))
		}
	}
	if len(scheduledAlerts) > 0 {
		datum.ScheduledAlerts = &scheduledAlerts
	}

	datum.Payload = data.NewBlob()
	if device.UDI != nil {
		(*datum.Payload)["udi"] = *device.UDI
	}
	if device.TransmitterGeneration != nil {
		(*datum.Payload)["transmitterGeneration"] = *device.TransmitterGeneration
	}
	if device.DisplayDevice != nil {
		(*datum.Payload)["displayDevice"] = *device.DisplayDevice
	}
	if device.SoftwareVersion != nil {
		(*datum.Payload)["softwareVersion"] = *device.SoftwareVersion
	}
	if device.SoftwareNumber != nil {
		(*datum.Payload)["softwareNumber"] = *device.SoftwareNumber
	}
	if device.Language != nil {
		(*datum.Payload)["language"] = *device.Language
	}
	if device.IsMmolDisplayMode != nil {
		(*datum.Payload)["isMmolDisplayMode"] = *device.IsMmolDisplayMode
	}
	if device.IsBlindedMode != nil {
		(*datum.Payload)["isBlindedMode"] = *device.IsBlindedMode
	}
	if device.Is24HourMode != nil {
		(*datum.Payload)["is24HourMode"] = *device.Is24HourMode
	}
	if device.DisplayTimeOffset != nil {
		(*datum.Payload)["displayTimeOffset"] = *device.DisplayTimeOffset
	}
	if device.SystemTimeOffset != nil {
		(*datum.Payload)["systemTimeOffset"] = *device.SystemTimeOffset
	}

	datum.Time = pointer.FromString(device.LastUploadDate.Format(dataTypes.TimeFormat))
	return datum
}

func translateAlertScheduleToScheduledAlert(alertSchedule *dexcom.AlertSchedule) *dataTypesSettingsCgm.ScheduledAlert {
	scheduledAlert := dataTypesSettingsCgm.NewScheduledAlert()
	scheduledAlert.Name = pointer.CloneString(alertSchedule.AlertScheduleSettings.Name)
	scheduledAlert.Days = translateAlertScheduleSettingsDaysOfWeekToScheduledAlertDays(alertSchedule.AlertScheduleSettings.DaysOfWeek)
	scheduledAlert.Start = translateAlertScheduleSettingsTimeToScheduledAlertTime(alertSchedule.AlertScheduleSettings.StartTime)
	scheduledAlert.End = translateAlertScheduleSettingsTimeToScheduledAlertTime(alertSchedule.AlertScheduleSettings.EndTime)
	scheduledAlert.Alerts = translateAlertSettingsToAlerts(alertSchedule.AlertScheduleSettings.Enabled, alertSchedule.AlertSettings)
	return scheduledAlert
}

func translateAlertScheduleSettingsDaysOfWeekToScheduledAlertDays(daysOfWeek *[]string) *[]string {
	if daysOfWeek == nil {
		return nil
	}
	days := []string{}
	for _, dayOfWeek := range *daysOfWeek {
		days = append(days, translateAlertScheduleSettingsDayOfWeekToScheduledAlertDay(dayOfWeek))
	}
	return &days
}

func translateAlertScheduleSettingsDayOfWeekToScheduledAlertDay(dayOfWeek string) string {
	switch dayOfWeek {
	case dexcom.AlertScheduleSettingsDaySunday:
		return dataTypesSettingsCgm.ScheduledAlertDaysSunday
	case dexcom.AlertScheduleSettingsDayMonday:
		return dataTypesSettingsCgm.ScheduledAlertDaysMonday
	case dexcom.AlertScheduleSettingsDayTuesday:
		return dataTypesSettingsCgm.ScheduledAlertDaysTuesday
	case dexcom.AlertScheduleSettingsDayWednesday:
		return dataTypesSettingsCgm.ScheduledAlertDaysWednesday
	case dexcom.AlertScheduleSettingsDayThursday:
		return dataTypesSettingsCgm.ScheduledAlertDaysThursday
	case dexcom.AlertScheduleSettingsDayFriday:
		return dataTypesSettingsCgm.ScheduledAlertDaysFriday
	case dexcom.AlertScheduleSettingsDaySaturday:
		return dataTypesSettingsCgm.ScheduledAlertDaysSaturday
	}
	return ""
}

func translateAlertScheduleSettingsTimeToScheduledAlertTime(tm *string) *int {
	if tm == nil {
		return nil
	}
	hour, minute, ok := dexcom.ParseAlertScheduleSettingsTime(*tm)
	if !ok {
		return nil
	}
	return pointer.FromInt((((hour * 60) + minute) * 60) * 1000)
}

func translateAlertSettingsToAlerts(enabled *bool, alertSettings *dexcom.AlertSettings) *dataTypesSettingsCgm.Alerts {
	alerts := dataTypesSettingsCgm.NewAlerts()
	alerts.Enabled = pointer.CloneBool(enabled)
	for _, alertSetting := range *alertSettings {
		var snooze *dataTypesSettingsCgm.Snooze
		if alertSetting.Snooze != nil {
			snooze = dataTypesSettingsCgm.NewSnooze()
			snooze.Duration = pointer.FromFloat64(float64(*alertSetting.Snooze))
			snooze.Units = pointer.FromString(dataTypesSettingsCgm.SnoozeUnitsMinutes)
		}

		switch *alertSetting.AlertName {
		case dexcom.AlertSettingAlertNameFall:
			alerts.Fall = dataTypesSettingsCgm.NewFallAlert()
			alerts.Fall.Enabled = pointer.CloneBool(alertSetting.Enabled)
			alerts.Fall.Snooze = snooze
			alerts.Fall.Rate = pointer.CloneFloat64(alertSetting.Value)
			alerts.Fall.Units = translateAlertSettingUnitToRateAlertUnits(alertSetting.Unit)
		case dexcom.AlertSettingAlertNameHigh:
			alerts.High = dataTypesSettingsCgm.NewHighAlert()
			alerts.High.Enabled = pointer.CloneBool(alertSetting.Enabled)
			alerts.High.Snooze = snooze
			alerts.High.Level = pointer.CloneFloat64(alertSetting.Value)
			alerts.High.Units = translateAlertSettingUnitToLevelAlertUnits(alertSetting.Unit)
		case dexcom.AlertSettingAlertNameLow:
			alerts.Low = dataTypesSettingsCgm.NewLowAlert()
			alerts.Low.Enabled = pointer.CloneBool(alertSetting.Enabled)
			alerts.Low.Snooze = snooze
			alerts.Low.Level = pointer.CloneFloat64(alertSetting.Value)
			alerts.Low.Units = translateAlertSettingUnitToLevelAlertUnits(alertSetting.Unit)
		case dexcom.AlertSettingAlertNameNoReadings:
			alerts.NoData = dataTypesSettingsCgm.NewNoDataAlert()
			alerts.NoData.Enabled = pointer.CloneBool(alertSetting.Enabled)
			alerts.NoData.Snooze = snooze
			alerts.NoData.Duration = pointer.CloneFloat64(alertSetting.Value)
			alerts.NoData.Units = translateAlertSettingUnitToDurationAlertUnits(alertSetting.Unit)
		case dexcom.AlertSettingAlertNameOutOfRange:
			alerts.OutOfRange = dataTypesSettingsCgm.NewOutOfRangeAlert()
			alerts.OutOfRange.Enabled = pointer.CloneBool(alertSetting.Enabled)
			alerts.OutOfRange.Snooze = snooze
			alerts.OutOfRange.Duration = pointer.CloneFloat64(alertSetting.Value)
			alerts.OutOfRange.Units = translateAlertSettingUnitToDurationAlertUnits(alertSetting.Unit)
		case dexcom.AlertSettingAlertNameRise:
			alerts.Rise = dataTypesSettingsCgm.NewRiseAlert()
			alerts.Rise.Enabled = pointer.CloneBool(alertSetting.Enabled)
			alerts.Rise.Snooze = snooze
			alerts.Rise.Rate = pointer.CloneFloat64(alertSetting.Value)
			alerts.Rise.Units = translateAlertSettingUnitToRateAlertUnits(alertSetting.Unit)
		case dexcom.AlertSettingAlertNameUrgentLow:
			alerts.UrgentLow = dataTypesSettingsCgm.NewUrgentLowAlert()
			alerts.UrgentLow.Enabled = pointer.CloneBool(alertSetting.Enabled)
			alerts.UrgentLow.Snooze = snooze
			alerts.UrgentLow.Level = pointer.CloneFloat64(alertSetting.Value)
			alerts.UrgentLow.Units = translateAlertSettingUnitToLevelAlertUnits(alertSetting.Unit)
		case dexcom.AlertSettingAlertNameUrgentLowSoon:
			alerts.UrgentLowPredicted = dataTypesSettingsCgm.NewUrgentLowAlert()
			alerts.UrgentLowPredicted.Enabled = pointer.CloneBool(alertSetting.Enabled)
			alerts.UrgentLowPredicted.Snooze = snooze
			alerts.UrgentLowPredicted.Level = pointer.CloneFloat64(alertSetting.Value)
			alerts.UrgentLowPredicted.Units = translateAlertSettingUnitToLevelAlertUnits(alertSetting.Unit)
		}
	}
	return alerts
}

func translateAlertSettingUnitToDurationAlertUnits(unit *string) *string {
	if unit != nil {
		switch *unit {
		case dexcom.AlertSettingUnitMinutes:
			return pointer.FromString(dataTypesSettingsCgm.DurationAlertUnitsMinutes)
		}
	}
	return nil
}

func translateAlertSettingUnitToLevelAlertUnits(unit *string) *string {
	if unit != nil {
		switch *unit {
		case dexcom.AlertSettingUnitMgdL:
			return pointer.FromString(dataTypesSettingsCgm.LevelAlertUnitsMgdL)
		}
	}
	return nil
}

func translateAlertSettingUnitToRateAlertUnits(unit *string) *string {
	if unit != nil {
		switch *unit {
		case dexcom.AlertSettingUnitMgdLMinute:
			return pointer.FromString(dataTypesSettingsCgm.RateAlertUnitsMgdLMinute)
		}
	}
	return nil
}

func translateEGVToDatum(egv *dexcom.EGV, unit *string, rateUnit *string) data.Datum {
	datum := dataTypesBloodGlucoseContinuous.New()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = nil
	datum.GUID = nil

	datum.Value = pointer.CloneFloat64(egv.Value)
	datum.Units = pointer.CloneString(unit)
	datum.Payload = data.NewBlob()
	if egv.RealTimeValue != nil {
		(*datum.Payload)["realTimeValue"] = *egv.RealTimeValue
	}
	if egv.SmoothedValue != nil {
		(*datum.Payload)["smoothedValue"] = *egv.SmoothedValue
	}
	if egv.Status != nil {
		(*datum.Payload)["status"] = *egv.Status
	}
	if egv.Trend != nil {
		(*datum.Payload)["trend"] = *egv.Trend
	}
	if egv.TrendRate != nil {
		(*datum.Payload)["trendRate"] = *egv.TrendRate
		(*datum.Payload)["trendRateUnits"] = *rateUnit
	}
	if egv.TransmitterID != nil {
		(*datum.Payload)["transmitterId"] = *egv.TransmitterID
	}
	if egv.TransmitterTicks != nil {
		(*datum.Payload)["transmitterTicks"] = *egv.TransmitterTicks
	}

	switch *datum.Units {
	case dexcom.EGVUnitMgdL:
		if *datum.Value < dexcom.EGVValuePinnedMgdLMinimum {
			datum.Value = pointer.FromFloat64(dexcom.EGVValuePinnedMgdLMinimum - 1)
			datum.Annotations = &data.BlobArray{{
				"code":      "bg/out-of-range",
				"value":     "low",
				"threshold": dexcom.EGVValuePinnedMgdLMinimum,
			}}
		} else if *datum.Value > dexcom.EGVValuePinnedMgdLMaximum {
			datum.Value = pointer.FromFloat64(dexcom.EGVValuePinnedMgdLMaximum + 1)
			datum.Annotations = &data.BlobArray{{
				"code":      "bg/out-of-range",
				"value":     "high",
				"threshold": dexcom.EGVValuePinnedMgdLMaximum,
			}}
		}
	}

	translateTime(egv.SystemTime, egv.DisplayTime, &datum.Base)
	return datum
}

func translateEventCarbsToDatum(event *dexcom.Event) data.Datum {
	datum := dataTypesFood.New()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = nil
	datum.GUID = nil

	if event.Value != nil && event.Unit != nil {
		datum.Nutrition = &dataTypesFood.Nutrition{
			Carbohydrate: &dataTypesFood.Carbohydrate{
				Net:   pointer.CloneFloat64(event.Value),
				Units: pointer.CloneString(event.Unit),
			},
		}
	}
	if event.ID != nil {
		datum.Origin = &dataTypesCommonOrigin.Origin{ID: pointer.CloneString(event.ID)}
	}

	translateTime(event.SystemTime, event.DisplayTime, &datum.Base)
	return datum
}

func translateEventExerciseToDatum(event *dexcom.Event) data.Datum {
	datum := dataTypesActivityPhysical.New()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = nil
	datum.GUID = nil

	if event.SubType != nil {
		switch *event.SubType {
		case dexcom.EventSubTypeExerciseLight:
			datum.ReportedIntensity = pointer.FromString(dataTypesActivityPhysical.ReportedIntensityLow)
		case dexcom.EventSubTypeExerciseMedium:
			datum.ReportedIntensity = pointer.FromString(dataTypesActivityPhysical.ReportedIntensityMedium)
		case dexcom.EventSubTypeExerciseHeavy:
			datum.ReportedIntensity = pointer.FromString(dataTypesActivityPhysical.ReportedIntensityHigh)
		}
	}
	if event.Value != nil && event.Unit != nil {
		datum.Duration = &dataTypesActivityPhysical.Duration{
			Units: pointer.CloneString(event.Unit),
			Value: pointer.CloneFloat64(event.Value),
		}
	}
	if event.ID != nil {
		datum.Origin = &dataTypesCommonOrigin.Origin{ID: pointer.CloneString(event.ID)}
	}

	translateTime(event.SystemTime, event.DisplayTime, &datum.Base)
	return datum
}

func translateEventHealthToDatum(event *dexcom.Event) data.Datum {
	datum := dataTypesStateReported.New()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = nil
	datum.GUID = nil

	if event.SubType != nil {
		switch *event.SubType {
		case dexcom.EventSubTypeHealthIllness:
			datum.States = &dataTypesStateReported.StateArray{{State: pointer.FromString(dataTypesStateReported.StateStateIllness)}}
		case dexcom.EventSubTypeHealthStress:
			datum.States = &dataTypesStateReported.StateArray{{State: pointer.FromString(dataTypesStateReported.StateStateStress)}}
		case dexcom.EventSubTypeHealthHighSymptoms:
			datum.States = &dataTypesStateReported.StateArray{{State: pointer.FromString(dataTypesStateReported.StateStateHyperglycemiaSymptoms)}}
		case dexcom.EventSubTypeHealthLowSymptoms:
			datum.States = &dataTypesStateReported.StateArray{{State: pointer.FromString(dataTypesStateReported.StateStateHypoglycemiaSymptoms)}}
		case dexcom.EventSubTypeHealthCycle:
			datum.States = &dataTypesStateReported.StateArray{{State: pointer.FromString(dataTypesStateReported.StateStateCycle)}}
		case dexcom.EventSubTypeHealthAlcohol:
			datum.States = &dataTypesStateReported.StateArray{{State: pointer.FromString(dataTypesStateReported.StateStateAlcohol)}}
		}
	}
	if event.ID != nil {
		datum.Origin = &dataTypesCommonOrigin.Origin{ID: pointer.CloneString(event.ID)}
	}

	translateTime(event.SystemTime, event.DisplayTime, &datum.Base)
	return datum
}

func translateEventInsulinToDatum(event *dexcom.Event) data.Datum {
	datum := dataTypesInsulin.New()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = nil
	datum.GUID = nil

	if event.SubType != nil {
		switch *event.SubType {
		case dexcom.EventSubTypeInsulinFastActing:
			datum.Formulation = &dataTypesInsulin.Formulation{Simple: &dataTypesInsulin.Simple{ActingType: pointer.FromString(dataTypesInsulin.SimpleActingTypeRapid)}}
		case dexcom.EventSubTypeInsulinLongActing:
			datum.Formulation = &dataTypesInsulin.Formulation{Simple: &dataTypesInsulin.Simple{ActingType: pointer.FromString(dataTypesInsulin.SimpleActingTypeLong)}}
		}
	}
	if event.Value != nil && event.Unit != nil {
		datum.Dose = &dataTypesInsulin.Dose{
			Total: pointer.CloneFloat64(event.Value),
			Units: pointer.FromString(dataTypesInsulin.DoseUnitsUnits),
		}
	}
	if event.ID != nil {
		datum.Origin = &dataTypesCommonOrigin.Origin{ID: pointer.CloneString(event.ID)}
	}

	translateTime(event.SystemTime, event.DisplayTime, &datum.Base)
	return datum
}
