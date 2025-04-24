package fetch

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesActivityPhysical "github.com/tidepool-org/platform/data/types/activity/physical"
	dataTypesAlert "github.com/tidepool-org/platform/data/types/alert"
	dataTypesBloodGlucoseContinuous "github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	dataTypesBloodGlucoseSelfMonitored "github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	dataTypesDeviceCalibration "github.com/tidepool-org/platform/data/types/device/calibration"
	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	dataTypesInsulin "github.com/tidepool-org/platform/data/types/insulin"
	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	dataTypesStateReported "github.com/tidepool-org/platform/data/types/state/reported"
	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/origin"
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

// Expectations:
// - systemTime must not be nil
// - displayTime can be nil (systemTime used, if so)
// - datum must not be nil
func TranslateTime(ctx context.Context, systemTime *dexcom.Time, displayTime *dexcom.Time, datum *dataTypes.Base) {
	var clockDriftOffsetDuration time.Duration
	var conversionOffsetDuration time.Duration
	var timeZoneOffsetDuration time.Duration

	// Get system time in UTC
	systemTimeUTC := systemTime.UTC()

	// Update datum
	datum.Time = pointer.FromTime(systemTimeUTC)
	if datum.Payload == nil {
		datum.Payload = metadata.NewMetadata()
	}
	datum.Payload.Set("systemTime", systemTime) // Original system time

	// If no display time, then no other calculations can be made
	if displayTime == nil {
		log.LoggerFromContext(ctx).Warn("Display time missing")
		return
	}

	delta := displayTime.Sub(systemTimeUTC)
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

	// If the display time zone was parsed, then force the time zone offset to match
	if displayTime.ZoneParsed() {

		// Apply any current time zone offset to the conversion offset
		conversionOffsetDuration += timeZoneOffsetDuration

		// Force time zone offset to what is specified in the display time
		_, displayTimeZoneOffset := displayTime.Zone()
		timeZoneOffsetDuration = time.Duration(displayTimeZoneOffset) * time.Second
	}

	// Update datum
	datum.DeviceTime = pointer.FromString(displayTime.Format(dataTypes.DeviceTimeFormat))
	datum.TimeZoneOffset = pointer.FromInt(int(timeZoneOffsetDuration / time.Minute))
	if clockDriftOffsetDuration != 0 {
		datum.ClockDriftOffset = pointer.FromInt(int(clockDriftOffsetDuration / time.Millisecond))
	}
	if conversionOffsetDuration != 0 {
		datum.ConversionOffset = pointer.FromInt(int(conversionOffsetDuration / time.Millisecond))
	}
	datum.Payload.Set("displayTime", displayTime) // Original display time
}

func translateCalibrationToDatum(ctx context.Context, calibration *dexcom.Calibration) data.Datum {
	datum := dataTypesDeviceCalibration.New()

	datum.DeviceID = TranslateDeviceIDFromTransmitter(calibration.TransmitterGeneration, calibration.TransmitterID)
	datum.Value = pointer.CloneFloat64(calibration.Value)
	datum.Units = pointer.CloneString(calibration.Unit)
	datum.Payload = metadata.NewMetadata()

	if calibration.TransmitterID != nil {
		(*datum.Payload)["transmitterId"] = *calibration.TransmitterID
	}
	if calibration.DisplayDevice != nil {
		(*datum.Payload)["displayDevice"] = *calibration.DisplayDevice
	}
	if calibration.TransmitterGeneration != nil {
		(*datum.Payload)["transmitterGeneration"] = *calibration.TransmitterGeneration
	}
	if calibration.TransmitterTicks != nil {
		(*datum.Payload)["transmitterTicks"] = *calibration.TransmitterTicks
	}
	if calibration.RecordID != nil {
		datum.Origin = &origin.Origin{ID: pointer.CloneString(calibration.RecordID)}
	}
	TranslateTime(ctx, calibration.SystemTime, calibration.DisplayTime, &datum.Base)
	return datum
}

func translateDeviceToDatum(_ context.Context, device *dexcom.Device) data.Datum {
	datum := dataTypesSettingsCgm.New()

	datum.DeviceID = TranslateDeviceIDFromTransmitter(device.TransmitterGeneration, device.TransmitterID)
	datum.Manufacturers = pointer.FromStringArray([]string{"Dexcom"})
	datum.TransmitterID = pointer.CloneString(device.TransmitterID)
	//TODO: potentially not true in the future. Currently the v3 API returns only MgdL but it does also have MmolL as valid units although it doesn't return them
	datum.Units = pointer.FromString(dataBloodGlucose.MgdL)

	defaultAlertSchedule := device.AlertSchedules.Default()
	if defaultAlertSchedule != nil {
		datum.DefaultAlerts = translateAlertSettingsToAlerts(defaultAlertSchedule.AlertScheduleSettings.IsEnabled, defaultAlertSchedule.AlertSettings)
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
	for _, alertSchedule := range *device.AlertSchedules {
		if alertSchedule != defaultAlertSchedule {
			scheduledAlerts = append(scheduledAlerts, translateAlertScheduleToScheduledAlert(alertSchedule))
		}
	}
	if len(scheduledAlerts) > 0 {
		datum.ScheduledAlerts = &scheduledAlerts
	}

	datum.Payload = metadata.NewMetadata()

	if device.TransmitterGeneration != nil {
		(*datum.Payload)["transmitterGeneration"] = *device.TransmitterGeneration
	}
	if device.DisplayDevice != nil {
		(*datum.Payload)["displayDevice"] = *device.DisplayDevice
	}

	datum.Time = pointer.FromTime(device.LastUploadDate.Time)
	return datum
}

func translateAlertScheduleToScheduledAlert(alertSchedule *dexcom.AlertSchedule) *dataTypesSettingsCgm.ScheduledAlert {
	scheduledAlert := dataTypesSettingsCgm.NewScheduledAlert()
	scheduledAlert.Name = pointer.CloneString(alertSchedule.AlertScheduleSettings.AlertScheduleName)
	scheduledAlert.Days = translateAlertScheduleSettingsDaysOfWeekToScheduledAlertDays(alertSchedule.AlertScheduleSettings.DaysOfWeek)
	scheduledAlert.Start = translateAlertScheduleSettingsTimeToScheduledAlertTime(alertSchedule.AlertScheduleSettings.StartTime)
	scheduledAlert.End = translateAlertScheduleSettingsTimeToScheduledAlertTime(alertSchedule.AlertScheduleSettings.EndTime)
	scheduledAlert.Alerts = translateAlertSettingsToAlerts(alertSchedule.AlertScheduleSettings.IsEnabled, alertSchedule.AlertSettings)
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
	hour, minute, err := dexcom.ParseAlertScheduleSettingsTimeHoursAndMinutes(*tm)
	if err != nil {
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
			return pointer.FromString(dataBloodGlucose.MgdLMinute)
		}
	}
	return nil
}

func translateAlertToDatum(ctx context.Context, alert *dexcom.Alert, version *string) data.Datum {
	datum := dataTypesAlert.New()

	datum.DeviceID = TranslateDeviceIDFromTransmitter(alert.TransmitterGeneration, alert.TransmitterID)
	datum.Payload = metadata.NewMetadata()

	if alert.AlertState != nil {
		(*datum.Payload)["alertState"] = *alert.AlertState
	}
	if alert.TransmitterID != nil {
		(*datum.Payload)["transmitterId"] = *alert.TransmitterID
	}
	if alert.TransmitterGeneration != nil {
		(*datum.Payload)["transmitterGeneration"] = *alert.TransmitterGeneration
	}
	if alert.DisplayDevice != nil {
		(*datum.Payload)["displayDevice"] = *alert.DisplayDevice
	}
	if version != nil {
		(*datum.Payload)["version"] = *version
	}

	if alert.RecordID != nil {
		datum.Origin = &origin.Origin{ID: pointer.CloneString(alert.RecordID)}
	}
	datum.IssuedTime = alert.DisplayTime.Raw()
	datum.Name = pointer.CloneString(alert.AlertName)
	TranslateTime(ctx, alert.SystemTime, alert.DisplayTime, &datum.Base)
	return datum
}

func translateEGVToDatum(ctx context.Context, egv *dexcom.EGV) data.Datum {
	datum := dataTypesBloodGlucoseContinuous.New()

	datum.DeviceID = TranslateDeviceIDFromTransmitter(egv.TransmitterGeneration, egv.TransmitterID)
	datum.Value = pointer.CloneFloat64(egv.Value)
	datum.Units = pointer.CloneString(egv.Unit)
	datum.Payload = metadata.NewMetadata()

	if egv.RateUnit != nil && egv.TrendRate != nil {
		switch *egv.RateUnit {
		case dexcom.EGVRateUnitMmolLMinute:
			datum.TrendRateUnits = pointer.FromString(dataBloodGlucose.MmolLMinute)
			datum.TrendRate = egv.TrendRate
		case dexcom.EGVRateUnitMgdLMinute:
			datum.TrendRateUnits = pointer.FromString(dataBloodGlucose.MgdLMinute)
			datum.TrendRate = egv.TrendRate
		case dexcom.EGVRateUnitUnknown:
			// NOP
		}
	}

	if egv.Trend != nil {
		switch *egv.Trend {
		case dexcom.EGVTrendDoubleUp:
			datum.Trend = pointer.FromString(dataTypesBloodGlucoseContinuous.RapidRise)
		case dexcom.EGVTrendSingleUp:
			datum.Trend = pointer.FromString(dataTypesBloodGlucoseContinuous.ModerateRise)
		case dexcom.EGVTrendFortyFiveUp:
			datum.Trend = pointer.FromString(dataTypesBloodGlucoseContinuous.SlowRise)
		case dexcom.EGVTrendFlat:
			datum.Trend = pointer.FromString(dataTypesBloodGlucoseContinuous.ConstantRate)
		case dexcom.EGVTrendFortyFiveDown:
			datum.Trend = pointer.FromString(dataTypesBloodGlucoseContinuous.SlowFall)
		case dexcom.EGVTrendSingleDown:
			datum.Trend = pointer.FromString(dataTypesBloodGlucoseContinuous.ModerateFall)
		case dexcom.EGVTrendDoubleDown:
			datum.Trend = pointer.FromString(dataTypesBloodGlucoseContinuous.RapidFall)
		case dexcom.EGVTrendUnknown, dexcom.EGVTrendNone, dexcom.EGVTrendNotComputable, dexcom.EGVTrendRateOutOfRange:
			// NOP
		}
	}

	if egv.Status != nil {
		(*datum.Payload)["status"] = *egv.Status
	}
	if egv.Trend != nil {
		(*datum.Payload)["trend"] = *egv.Trend
	}
	if egv.TrendRate != nil {
		(*datum.Payload)["trendRate"] = *egv.TrendRate
		(*datum.Payload)["trendRateUnits"] = *egv.RateUnit
	}
	if egv.TransmitterID != nil {
		(*datum.Payload)["transmitterId"] = *egv.TransmitterID
	}
	if egv.TransmitterTicks != nil {
		(*datum.Payload)["transmitterTicks"] = *egv.TransmitterTicks
	}
	if egv.DisplayDevice != nil {
		(*datum.Payload)["displayDevice"] = *egv.DisplayDevice
	}
	if egv.TransmitterGeneration != nil {
		(*datum.Payload)["transmitterGeneration"] = *egv.TransmitterGeneration
	}

	switch *datum.Units {
	case dexcom.EGVUnitMgdL:
		if *datum.Value < dexcom.EGVValuePinnedMgdLMinimum {
			datum.Value = pointer.FromFloat64(dexcom.EGVValuePinnedMgdLMinimum - 1)
			datum.Annotations = &metadata.MetadataArray{{
				"code":      "bg/out-of-range",
				"value":     "low",
				"threshold": dexcom.EGVValuePinnedMgdLMinimum,
			}}
		} else if *datum.Value > dexcom.EGVValuePinnedMgdLMaximum {
			datum.Value = pointer.FromFloat64(dexcom.EGVValuePinnedMgdLMaximum + 1)
			datum.Annotations = &metadata.MetadataArray{{
				"code":      "bg/out-of-range",
				"value":     "high",
				"threshold": dexcom.EGVValuePinnedMgdLMaximum,
			}}
		}
	case dexcom.EGVUnitMmolL:
		if *datum.Value < dexcom.EGVValuePinnedMmolLMinimum {
			datum.Value = pointer.FromFloat64(dexcom.EGVValuePinnedMmolLMinimum - 0.1)
			datum.Annotations = &metadata.MetadataArray{{
				"code":      "bg/out-of-range",
				"value":     "low",
				"threshold": dexcom.EGVValuePinnedMmolLMinimum,
			}}
		} else if *datum.Value > dexcom.EGVValuePinnedMmolLMaximum {
			datum.Value = pointer.FromFloat64(dexcom.EGVValuePinnedMmolLMaximum + 0.1)
			datum.Annotations = &metadata.MetadataArray{{
				"code":      "bg/out-of-range",
				"value":     "high",
				"threshold": dexcom.EGVValuePinnedMmolLMaximum,
			}}
		}
	}
	if egv.RecordID != nil {
		datum.Origin = &origin.Origin{ID: pointer.CloneString(egv.RecordID)}
	}
	TranslateTime(ctx, egv.SystemTime, egv.DisplayTime, &datum.Base)
	return datum
}

func translateEventCarbsToDatum(ctx context.Context, event *dexcom.Event) data.Datum {
	datum := dataTypesFood.New()

	datum.DeviceID = TranslateDeviceIDFromTransmitter(event.TransmitterGeneration, event.TransmitterID)
	if event.Value != nil && event.Unit != nil {
		floatVal, _ := strconv.ParseFloat(*event.Value, 64)
		datum.Nutrition = &dataTypesFood.Nutrition{
			Carbohydrate: &dataTypesFood.Carbohydrate{
				Net:   pointer.CloneFloat64(&floatVal),
				Units: pointer.CloneString(event.Unit),
			},
		}
	}
	if event.RecordID != nil {
		datum.Origin = &origin.Origin{ID: pointer.CloneString(event.RecordID)}
	}

	TranslateTime(ctx, event.SystemTime, event.DisplayTime, &datum.Base)
	return datum
}

func translateEventExerciseToDatum(ctx context.Context, event *dexcom.Event) data.Datum {
	datum := dataTypesActivityPhysical.New()

	datum.DeviceID = TranslateDeviceIDFromTransmitter(event.TransmitterGeneration, event.TransmitterID)
	if event.EventSubType != nil {
		switch *event.EventSubType {
		case dexcom.EventSubTypeExerciseLight:
			datum.ReportedIntensity = pointer.FromString(dataTypesActivityPhysical.ReportedIntensityLow)
		case dexcom.EventSubTypeExerciseMedium:
			datum.ReportedIntensity = pointer.FromString(dataTypesActivityPhysical.ReportedIntensityMedium)
		case dexcom.EventSubTypeExerciseHeavy:
			datum.ReportedIntensity = pointer.FromString(dataTypesActivityPhysical.ReportedIntensityHigh)
		}
	}
	if event.Value != nil && event.Unit != nil {
		floatVal, err := strconv.ParseFloat(*event.Value, 64)
		if err == nil {
			datum.Duration = &dataTypesActivityPhysical.Duration{
				Units: pointer.CloneString(event.Unit),
				Value: pointer.CloneFloat64(&floatVal),
			}
		}
	}
	if event.RecordID != nil {
		datum.Origin = &origin.Origin{ID: pointer.CloneString(event.RecordID)}
	}

	TranslateTime(ctx, event.SystemTime, event.DisplayTime, &datum.Base)
	return datum
}

func translateEventHealthToDatum(ctx context.Context, event *dexcom.Event) data.Datum {
	datum := dataTypesStateReported.New()

	datum.DeviceID = TranslateDeviceIDFromTransmitter(event.TransmitterGeneration, event.TransmitterID)
	if event.EventSubType != nil {
		switch *event.EventSubType {
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
	if event.RecordID != nil {
		datum.Origin = &origin.Origin{ID: pointer.CloneString(event.RecordID)}
	}

	TranslateTime(ctx, event.SystemTime, event.DisplayTime, &datum.Base)
	return datum
}

func translateEventInsulinToDatum(ctx context.Context, event *dexcom.Event) data.Datum {
	datum := dataTypesInsulin.New()

	datum.DeviceID = TranslateDeviceIDFromTransmitter(event.TransmitterGeneration, event.TransmitterID)
	if event.EventSubType != nil {
		switch *event.EventSubType {
		case dexcom.EventSubTypeInsulinFastActing:
			datum.Formulation = &dataTypesInsulin.Formulation{Simple: &dataTypesInsulin.Simple{ActingType: pointer.FromString(dataTypesInsulin.SimpleActingTypeRapid)}}
		case dexcom.EventSubTypeInsulinLongActing:
			datum.Formulation = &dataTypesInsulin.Formulation{Simple: &dataTypesInsulin.Simple{ActingType: pointer.FromString(dataTypesInsulin.SimpleActingTypeLong)}}
		}
	}
	if event.Value != nil && event.Unit != nil {
		floatVal, err := strconv.ParseFloat(*event.Value, 64)
		if err == nil {
			datum.Dose = &dataTypesInsulin.Dose{
				Total: pointer.CloneFloat64(&floatVal),
				Units: pointer.FromString(dataTypesInsulin.DoseUnitsUnits),
			}
		}
	}
	if event.RecordID != nil {
		datum.Origin = &origin.Origin{ID: pointer.CloneString(event.RecordID)}
	}

	TranslateTime(ctx, event.SystemTime, event.DisplayTime, &datum.Base)

	return datum
}

func translateEventBloodGlucoseToDatum(ctx context.Context, event *dexcom.Event) data.Datum {
	datum := dataTypesBloodGlucoseSelfMonitored.New()

	datum.DeviceID = TranslateDeviceIDFromTransmitter(event.TransmitterGeneration, event.TransmitterID)
	if event.Value != nil && event.Unit != nil {
		floatVal, err := strconv.ParseFloat(*event.Value, 64)
		if err == nil {
			datum.Value = pointer.CloneFloat64(&floatVal)
		}
		datum.Units = pointer.CloneString(event.Unit)
	}

	if event.RecordID != nil {
		datum.Origin = &origin.Origin{ID: pointer.CloneString(event.RecordID)}
	}

	TranslateTime(ctx, event.SystemTime, event.DisplayTime, &datum.Base)
	return datum
}

func translateEventNotesToDatum(ctx context.Context, event *dexcom.Event) data.Datum {
	datum := dataTypesStateReported.New()

	datum.DeviceID = TranslateDeviceIDFromTransmitter(event.TransmitterGeneration, event.TransmitterID)
	if event.EventSubType != nil {
		switch *event.EventSubType {
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
	if event.RecordID != nil {
		datum.Origin = &origin.Origin{ID: pointer.CloneString(event.RecordID)}
	}

	if event.Value != nil {
		datum.Notes = pointer.FromStringArray([]string{})
	}

	TranslateTime(ctx, event.SystemTime, event.DisplayTime, &datum.Base)
	return datum
}

func TranslateDeviceIDFromTransmitter(transmitterGeneration *string, transmitterID *string) *string {
	prefix := TranslateDeviceIDPrefixFromTransmitterGeneration(transmitterGeneration)
	if prefix == nil || transmitterID == nil || *transmitterID == "" {
		return nil
	}
	return pointer.FromString(strings.Join([]string{*prefix, *transmitterID}, "_"))
}

func TranslateDeviceIDPrefixFromTransmitterGeneration(transmitterGeneration *string) *string {
	if transmitterGeneration == nil {
		return nil
	}

	switch *transmitterGeneration {
	case dexcom.DeviceTransmitterGenerationUnknown:
		return pointer.FromString("Dexcom")
	case dexcom.DeviceTransmitterGenerationG4:
		return pointer.FromString("DexcomG4")
	case dexcom.DeviceTransmitterGenerationG5:
		return pointer.FromString("DexcomG5")
	case dexcom.DeviceTransmitterGenerationG6:
		return pointer.FromString("DexcomG6")
	case dexcom.DeviceTransmitterGenerationG6Pro:
		return pointer.FromString("DexcomG6Pro")
	case dexcom.DeviceTransmitterGenerationG6Plus:
		return pointer.FromString("DexcomG6Plus")
	case dexcom.DeviceTransmitterGenerationPro:
		return pointer.FromString("DexcomPro")
	case dexcom.DeviceTransmitterGenerationG7:
		return pointer.FromString("DexcomG7")
	default:
		return nil
	}
}
