package fetch

import (
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/activity/physical"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/device/calibration"
	"github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/data/types/state/reported"
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

const OffsetDuration = 30 * time.Minute // Duration between timezone offsets we scan for

const MaximumOffsets = (14 * time.Hour) / OffsetDuration  // Maximum timezone offset is +14:00
const MinimumOffsets = (-12 * time.Hour) / OffsetDuration // Minimum timezone offset is -12:00

const DailyDuration = 24 * time.Hour
const DailyOffsets = DailyDuration / OffsetDuration

func translateTime(systemTime time.Time, displayTime time.Time, datum *types.Base) {
	var clockDriftOffsetDuration time.Duration
	var conversionOffsetDuration time.Duration
	var timezoneOffsetDuration time.Duration

	delta := displayTime.Sub(systemTime)
	if delta > 0 {
		offsetCount := time.Duration((float64(delta) + float64(OffsetDuration)/2) / float64(OffsetDuration))
		clockDriftOffsetDuration = delta - offsetCount*OffsetDuration
		for offsetCount > MaximumOffsets {
			conversionOffsetDuration += DailyDuration
			offsetCount -= DailyOffsets
		}
		timezoneOffsetDuration = offsetCount * OffsetDuration
	} else if delta < 0 {
		offsetCount := time.Duration((float64(delta) - float64(OffsetDuration)/2) / float64(OffsetDuration))
		clockDriftOffsetDuration = delta - offsetCount*OffsetDuration
		for offsetCount < MinimumOffsets {
			conversionOffsetDuration -= DailyDuration
			offsetCount += DailyOffsets
		}
		timezoneOffsetDuration = offsetCount * OffsetDuration
	}

	datum.Time = pointer.String(systemTime.Format(types.TimeFormat))
	datum.DeviceTime = pointer.String(displayTime.Format(types.DeviceTimeFormat))
	datum.TimezoneOffset = pointer.Int(int(timezoneOffsetDuration / time.Minute))
	if clockDriftOffsetDuration != 0 {
		datum.ClockDriftOffset = pointer.Int(int(clockDriftOffsetDuration / time.Millisecond))
	}
	if conversionOffsetDuration != 0 {
		datum.ConversionOffset = pointer.Int(int(conversionOffsetDuration / time.Millisecond))
	}

	if datum.Payload == nil {
		datum.Payload = &map[string]interface{}{}
	}
	(*datum.Payload)["systemTime"] = systemTime
}

func translateCalibrationToDatum(c *dexcom.Calibration) data.Datum {
	datum := calibration.Init()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = ""
	datum.GUID = ""

	datum.Units = pointer.String(c.Unit)
	datum.Value = pointer.Float64(c.Value)
	datum.Payload = &map[string]interface{}{}
	if c.TransmitterID != nil {
		(*datum.Payload)["transmitterId"] = *c.TransmitterID
	}

	translateTime(c.SystemTime, c.DisplayTime, &datum.Base)
	return datum
}

func translateEGVToDatum(e *dexcom.EGV, unit string, rateUnit string) data.Datum {
	datum := continuous.Init()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = ""
	datum.GUID = ""

	datum.Value = pointer.Float64(e.Value)
	datum.Units = pointer.String(unit)
	datum.Payload = &map[string]interface{}{}
	if e.Status != nil {
		(*datum.Payload)["status"] = *e.Status
	}
	if e.Trend != nil {
		(*datum.Payload)["trend"] = *e.Trend
	}
	if e.TrendRate != nil {
		(*datum.Payload)["trendRate"] = *e.TrendRate
		(*datum.Payload)["trendRateUnits"] = rateUnit
	}
	if e.TransmitterID != nil {
		(*datum.Payload)["transmitterId"] = *e.TransmitterID
	}
	if e.TransmitterTicks != nil {
		(*datum.Payload)["transmitterTicks"] = *e.TransmitterTicks
	}

	switch unit {
	case dexcom.UnitMgdL:
		if e.Value < dexcom.EGVValueMinMgdL {
			datum.Annotations = &[]map[string]interface{}{{
				"code":      "bg/out-of-range",
				"value":     "low",
				"threshold": dexcom.EGVValueMinMgdL,
			}}
		} else if e.Value > dexcom.EGVValueMaxMgdL {
			datum.Annotations = &[]map[string]interface{}{{
				"code":      "bg/out-of-range",
				"value":     "high",
				"threshold": dexcom.EGVValueMaxMgdL,
			}}
		}
	case dexcom.UnitMmolL:
		// TODO: Add annotations
	}

	translateTime(e.SystemTime, e.DisplayTime, &datum.Base)
	return datum
}

func translateEventCarbsToDatum(e *dexcom.Event) data.Datum {
	datum := food.Init()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = ""
	datum.GUID = ""

	if e.Value != nil && e.Unit != nil {
		datum.Nutrition = &food.Nutrition{
			Carbohydrates: &food.Carbohydrates{
				Net:   pointer.Int(int(*e.Value)),
				Units: pointer.String(*e.Unit),
			},
		}
	}

	translateTime(e.SystemTime, e.DisplayTime, &datum.Base)
	return datum
}

func translateEventExerciseToDatum(e *dexcom.Event) data.Datum {
	datum := physical.Init()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = ""
	datum.GUID = ""

	if e.EventSubType != nil {
		switch *e.EventSubType {
		case dexcom.ExerciseLight:
			datum.ReportedIntensity = pointer.String(physical.ReportedIntensityLow)
		case dexcom.ExerciseMedium:
			datum.ReportedIntensity = pointer.String(physical.ReportedIntensityMedium)
		case dexcom.ExerciseHeavy:
			datum.ReportedIntensity = pointer.String(physical.ReportedIntensityHigh)
		}
	}
	if e.Value != nil && e.Unit != nil {
		datum.Duration = &physical.Duration{
			Value: pointer.Float64(*e.Value),
			Units: pointer.String(*e.Unit),
		}
	}

	translateTime(e.SystemTime, e.DisplayTime, &datum.Base)
	return datum
}

func translateEventHealthToDatum(e *dexcom.Event) data.Datum {
	datum := reported.Init()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = ""
	datum.GUID = ""

	if e.EventSubType != nil {
		switch *e.EventSubType {
		case dexcom.HealthIllness:
			datum.States = &[]*reported.State{{State: pointer.String(reported.StateIllness)}}
		case dexcom.HealthStress:
			datum.States = &[]*reported.State{{State: pointer.String(reported.StateStress)}}
		case dexcom.HealthHighSymptoms:
			datum.States = &[]*reported.State{{State: pointer.String(reported.StateHyperglycemiaSymptoms)}}
		case dexcom.HealthLowSymptoms:
			datum.States = &[]*reported.State{{State: pointer.String(reported.StateHypoglycemiaSymptoms)}}
		case dexcom.HealthCycle:
			datum.States = &[]*reported.State{{State: pointer.String(reported.StateCycle)}}
		case dexcom.HealthAlcohol:
			datum.States = &[]*reported.State{{State: pointer.String(reported.StateAlcohol)}}
		}
	}

	translateTime(e.SystemTime, e.DisplayTime, &datum.Base)
	return datum
}

func translateEventInsulinToDatum(e *dexcom.Event) data.Datum {
	datum := insulin.Init()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = ""
	datum.GUID = ""

	if e.Value != nil && e.Unit != nil {
		datum.Dose = &insulin.Dose{
			Total: pointer.Float64(*e.Value),
			Units: pointer.String(*e.Unit),
		}
	}

	translateTime(e.SystemTime, e.DisplayTime, &datum.Base)
	return datum
}
