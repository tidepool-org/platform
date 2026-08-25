package clinics

import (
	"strconv"
	"strings"
	"time"

	clinic "github.com/tidepool-org/clinic/client"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/summary/types"
)

// NewPatientSummary returns the summaries given as the patient summary document the clinic service
// accepts. A type given as nil is omitted from the document, which the clinic service leaves
// untouched, so each type is reported independently. Continuous summaries have no clinic
// representation.
func NewPatientSummary(cgm *types.CGMSummary, bgm *types.BGMSummary) *clinic.PatientSummaryV1 {
	patientSummary := &clinic.PatientSummaryV1{}
	if cgm != nil {
		patientSummary.CgmStats = &clinic.CgmStatsV1{
			Id:      pointer.FromString(cgm.ID.Hex()),
			Config:  exportSummaryConfig(cgm.Config),
			Dates:   exportSummaryDates(cgm.Dates),
			Periods: exportCGMPeriods(cgm.Periods),
		}
	}
	if bgm != nil {
		patientSummary.BgmStats = &clinic.BgmStatsV1{
			Id:      pointer.FromString(bgm.ID.Hex()),
			Config:  exportSummaryConfig(bgm.Config),
			Dates:   exportSummaryDates(bgm.Dates),
			Periods: exportBGMPeriods(bgm.Periods),
		}
	}
	return patientSummary
}

func exportSummaryConfig(config types.Config) clinic.SummaryConfigV1 {
	return clinic.SummaryConfigV1{
		SchemaVersion:            config.SchemaVersion,
		HighGlucoseThreshold:     config.HighGlucoseThreshold,
		VeryHighGlucoseThreshold: config.VeryHighGlucoseThreshold,
		LowGlucoseThreshold:      config.LowGlucoseThreshold,
		VeryLowGlucoseThreshold:  config.VeryLowGlucoseThreshold,
	}
}

func exportSummaryDates(dates types.Dates) clinic.SummaryDatesV1 {
	firstData := timeOrNil(dates.FirstData)
	lastData := timeOrNil(dates.LastData)
	lastUploadDate := timeOrNil(dates.LastUploadDate)

	// The reasons and the outdated mark are no longer maintained; the reasons are reported as empty
	// rather than omitted so that the clinic service clears any values reported before their removal
	return clinic.SummaryDatesV1{
		LastUpdatedDate:   timeOrNil(dates.LastUpdatedDate),
		LastUpdatedReason: pointer.FromAny([]string{}),
		OutdatedReason:    pointer.FromAny([]string{}),
		HasLastUploadDate: lastUploadDate != nil,
		LastUploadDate:    lastUploadDate,
		HasFirstData:      firstData != nil,
		FirstData:         firstData,
		HasLastData:       lastData != nil,
		LastData:          lastData,
	}
}

func timeOrNil(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return pointer.FromTime(value)
}

func exportCGMPeriods(periods *types.CGMPeriods) clinic.CgmPeriodsV1 {
	exported := clinic.CgmPeriodsV1{}
	if periods == nil {
		return exported
	}
	for name, period := range periods.GlucosePeriods {
		if days, ok := periodDays(name); ok && period != nil {
			exported[name] = exportCGMPeriod(period, days)
		}
	}
	return exported
}

func exportBGMPeriods(periods *types.BGMPeriods) clinic.BgmPeriodsV1 {
	exported := clinic.BgmPeriodsV1{}
	if periods == nil {
		return exported
	}
	for name, period := range periods.GlucosePeriods {
		if _, ok := periodDays(name); ok && period != nil {
			exported[name] = exportBGMPeriod(period)
		}
	}
	return exported
}

// periodDays reports the number of days of a period named 1d/7d/14d/30d
func periodDays(name string) (int, bool) {
	days, err := strconv.Atoi(strings.TrimSuffix(name, "d"))
	if !strings.HasSuffix(name, "d") || err != nil || days <= 0 {
		return 0, false
	}
	return days, true
}

func exportCGMPeriod(period *types.GlucosePeriod, days int) clinic.CgmPeriodV1 {
	delta := period.Delta
	if delta == nil {
		delta = &types.GlucosePeriod{}
	}

	exported := clinic.CgmPeriodV1{
		AverageDailyRecords:           pointer.FromFloat64(period.AverageDailyRecords),
		AverageDailyRecordsDelta:      pointer.FromFloat64(delta.AverageDailyRecords),
		DaysWithData:                  period.DaysWithData,
		DaysWithDataDelta:             delta.DaysWithData,
		HasAverageDailyRecords:        period.AverageDailyRecords != 0,
		HasTimeCGMUseMinutes:          period.Total.Minutes != 0,
		HasTimeCGMUseRecords:          period.Total.Records != 0,
		HasTimeInAnyHighMinutes:       period.AnyHigh.Minutes != 0,
		HasTimeInAnyHighRecords:       period.AnyHigh.Records != 0,
		HasTimeInAnyLowMinutes:        period.AnyLow.Minutes != 0,
		HasTimeInAnyLowRecords:        period.AnyLow.Records != 0,
		HasTimeInExtremeHighMinutes:   period.ExtremeHigh.Minutes != 0,
		HasTimeInExtremeHighRecords:   period.ExtremeHigh.Records != 0,
		HasTimeInHighMinutes:          period.High.Minutes != 0,
		HasTimeInHighRecords:          period.High.Records != 0,
		HasTimeInLowMinutes:           period.Low.Minutes != 0,
		HasTimeInLowRecords:           period.Low.Records != 0,
		HasTimeInTargetMinutes:        period.Target.Minutes != 0,
		HasTimeInTargetRecords:        period.Target.Records != 0,
		HasTimeInVeryHighMinutes:      period.VeryHigh.Minutes != 0,
		HasTimeInVeryHighRecords:      period.VeryHigh.Records != 0,
		HasTimeInVeryLowMinutes:       period.VeryLow.Minutes != 0,
		HasTimeInVeryLowRecords:       period.VeryLow.Records != 0,
		HasTotalRecords:               period.Total.Records != 0,
		HoursWithData:                 period.HoursWithData,
		HoursWithDataDelta:            delta.HoursWithData,
		TimeCGMUseMinutes:             pointer.FromInt(period.Total.Minutes),
		TimeCGMUseMinutesDelta:        pointer.FromInt(delta.Total.Minutes),
		TimeCGMUseRecords:             pointer.FromInt(period.Total.Records),
		TimeCGMUseRecordsDelta:        pointer.FromInt(delta.Total.Records),
		TimeInAnyHighMinutes:          pointer.FromInt(period.AnyHigh.Minutes),
		TimeInAnyHighMinutesDelta:     pointer.FromInt(delta.AnyHigh.Minutes),
		TimeInAnyHighRecords:          pointer.FromInt(period.AnyHigh.Records),
		TimeInAnyHighRecordsDelta:     pointer.FromInt(delta.AnyHigh.Records),
		TimeInAnyLowMinutes:           pointer.FromInt(period.AnyLow.Minutes),
		TimeInAnyLowMinutesDelta:      pointer.FromInt(delta.AnyLow.Minutes),
		TimeInAnyLowRecords:           pointer.FromInt(period.AnyLow.Records),
		TimeInAnyLowRecordsDelta:      pointer.FromInt(delta.AnyLow.Records),
		TimeInExtremeHighMinutes:      pointer.FromInt(period.ExtremeHigh.Minutes),
		TimeInExtremeHighMinutesDelta: pointer.FromInt(delta.ExtremeHigh.Minutes),
		TimeInExtremeHighRecords:      pointer.FromInt(period.ExtremeHigh.Records),
		TimeInExtremeHighRecordsDelta: pointer.FromInt(delta.ExtremeHigh.Records),
		TimeInHighMinutes:             pointer.FromInt(period.High.Minutes),
		TimeInHighMinutesDelta:        pointer.FromInt(delta.High.Minutes),
		TimeInHighRecords:             pointer.FromInt(period.High.Records),
		TimeInHighRecordsDelta:        pointer.FromInt(delta.High.Records),
		TimeInLowMinutes:              pointer.FromInt(period.Low.Minutes),
		TimeInLowMinutesDelta:         pointer.FromInt(delta.Low.Minutes),
		TimeInLowRecords:              pointer.FromInt(period.Low.Records),
		TimeInLowRecordsDelta:         pointer.FromInt(delta.Low.Records),
		TimeInTargetMinutes:           pointer.FromInt(period.Target.Minutes),
		TimeInTargetMinutesDelta:      pointer.FromInt(delta.Target.Minutes),
		TimeInTargetRecords:           pointer.FromInt(period.Target.Records),
		TimeInTargetRecordsDelta:      pointer.FromInt(delta.Target.Records),
		TimeInVeryHighMinutes:         pointer.FromInt(period.VeryHigh.Minutes),
		TimeInVeryHighMinutesDelta:    pointer.FromInt(delta.VeryHigh.Minutes),
		TimeInVeryHighRecords:         pointer.FromInt(period.VeryHigh.Records),
		TimeInVeryHighRecordsDelta:    pointer.FromInt(delta.VeryHigh.Records),
		TimeInVeryLowMinutes:          pointer.FromInt(period.VeryLow.Minutes),
		TimeInVeryLowMinutesDelta:     pointer.FromInt(delta.VeryLow.Minutes),
		TimeInVeryLowRecords:          pointer.FromInt(period.VeryLow.Records),
		TimeInVeryLowRecordsDelta:     pointer.FromInt(delta.VeryLow.Records),
		TotalRecords:                  pointer.FromInt(period.Total.Records),
		TotalRecordsDelta:             pointer.FromInt(delta.Total.Records),
		Min:                           period.Min,
		MinDelta:                      delta.Min,
		Max:                           period.Max,
		MaxDelta:                      delta.Max,
	}

	// reconstruct some previous period values for comparison later
	previousTotalRecords := period.Total.Records - delta.Total.Records
	previousCGMUsePercent := period.Total.Percent - delta.Total.Percent
	previousCGMUseMinutes := period.Total.Minutes - delta.Total.Minutes

	// The following provides concessions to allow patient list sorting and filtering according to
	// certain eligibility requirements, notably:
	// - TIR percent only is visible in the frontend if >1d of data, or 70% cgm use on single day metrics
	// - GMI requires >70% cgm use
	// - All percentages should be nil if 0 TotalRecords, as they would have been before schema v5
	// - All delta percentages should be nil if both periods do not fulfill their respective requirements above
	if period.Total.Records != 0 {
		exported.HasTimeCGMUsePercent = true
		exported.HasAverageGlucoseMmol = true
		exported.TimeCGMUsePercent = pointer.FromFloat64(period.Total.Percent)
		exported.AverageGlucoseMmol = pointer.FromFloat64(period.AverageGlucose)
		exported.StandardDeviation = period.StandardDeviation
		exported.CoefficientOfVariation = period.CoefficientOfVariation

		if previousTotalRecords != 0 {
			exported.TimeCGMUsePercentDelta = pointer.FromFloat64(delta.Total.Percent)
			exported.AverageGlucoseMmolDelta = pointer.FromFloat64(delta.AverageGlucose)
			exported.StandardDeviationDelta = delta.StandardDeviation
			exported.CoefficientOfVariationDelta = delta.CoefficientOfVariation
		}

		// if we are storing under 1d, apply 70% rule to TimeIn*
		// if we are storing over 1d, check for 24h cgm use
		if (days <= 1 && period.Total.Percent > 0.7) || (days > 1 && period.Total.Minutes > 1440) {
			exported.HasTimeInTargetPercent = true
			exported.TimeInTargetPercent = pointer.FromFloat64(period.Target.Percent)

			exported.HasTimeInLowPercent = true
			exported.TimeInLowPercent = pointer.FromFloat64(period.Low.Percent)

			exported.HasTimeInVeryLowPercent = true
			exported.TimeInVeryLowPercent = pointer.FromFloat64(period.VeryLow.Percent)

			exported.HasTimeInAnyLowPercent = true
			exported.TimeInAnyLowPercent = pointer.FromFloat64(period.AnyLow.Percent)

			exported.HasTimeInHighPercent = true
			exported.TimeInHighPercent = pointer.FromFloat64(period.High.Percent)

			exported.HasTimeInVeryHighPercent = true
			exported.TimeInVeryHighPercent = pointer.FromFloat64(period.VeryHigh.Percent)

			exported.HasTimeInExtremeHighPercent = true
			exported.TimeInExtremeHighPercent = pointer.FromFloat64(period.ExtremeHigh.Percent)

			exported.HasTimeInAnyHighPercent = true
			exported.TimeInAnyHighPercent = pointer.FromFloat64(period.AnyHigh.Percent)

			// add deltas if delta period fulfills requirements as well
			if (days <= 1 && previousCGMUsePercent > 0.7) || (days > 1 && previousCGMUseMinutes > 1440) {
				exported.TimeInTargetPercentDelta = pointer.FromFloat64(delta.Target.Percent)
				exported.TimeInLowPercentDelta = pointer.FromFloat64(delta.Low.Percent)
				exported.TimeInVeryLowPercentDelta = pointer.FromFloat64(delta.VeryLow.Percent)
				exported.TimeInAnyLowPercentDelta = pointer.FromFloat64(delta.AnyLow.Percent)
				exported.TimeInHighPercentDelta = pointer.FromFloat64(delta.High.Percent)
				exported.TimeInVeryHighPercentDelta = pointer.FromFloat64(delta.VeryHigh.Percent)
				exported.TimeInExtremeHighPercentDelta = pointer.FromFloat64(delta.ExtremeHigh.Percent)
				exported.TimeInAnyHighPercentDelta = pointer.FromFloat64(delta.AnyHigh.Percent)
			}
		}

		// GMI should only be present if CGM use % is >70% so that they are filtered to the bottom on GMI queries.
		if period.Total.Percent > 0.7 {
			exported.HasGlucoseManagementIndicator = true
			exported.GlucoseManagementIndicator = pointer.FromFloat64(period.GlucoseManagementIndicator)

			// add deltas if delta period fulfills requirements as well
			if previousCGMUsePercent > 0.7 {
				exported.GlucoseManagementIndicatorDelta = pointer.FromFloat64(delta.GlucoseManagementIndicator)
			}
		}
	}

	return exported
}

func exportBGMPeriod(period *types.GlucosePeriod) clinic.BgmPeriodV1 {
	delta := period.Delta
	if delta == nil {
		delta = &types.GlucosePeriod{}
	}

	exported := clinic.BgmPeriodV1{
		AverageDailyRecords:           pointer.FromFloat64(period.AverageDailyRecords),
		AverageDailyRecordsDelta:      pointer.FromFloat64(delta.AverageDailyRecords),
		DaysWithData:                  period.DaysWithData,
		DaysWithDataDelta:             delta.DaysWithData,
		HasAverageDailyRecords:        period.AverageDailyRecords != 0,
		HasTimeInAnyHighRecords:       period.AnyHigh.Records != 0,
		HasTimeInAnyLowRecords:        period.AnyLow.Records != 0,
		HasTimeInExtremeHighRecords:   period.ExtremeHigh.Records != 0,
		HasTimeInHighRecords:          period.High.Records != 0,
		HasTimeInLowRecords:           period.Low.Records != 0,
		HasTimeInTargetRecords:        period.Target.Records != 0,
		HasTimeInVeryHighRecords:      period.VeryHigh.Records != 0,
		HasTimeInVeryLowRecords:       period.VeryLow.Records != 0,
		HasTotalRecords:               period.Total.Records != 0,
		TimeInAnyHighRecords:          pointer.FromInt(period.AnyHigh.Records),
		TimeInAnyHighRecordsDelta:     pointer.FromInt(delta.AnyHigh.Records),
		TimeInAnyLowRecords:           pointer.FromInt(period.AnyLow.Records),
		TimeInAnyLowRecordsDelta:      pointer.FromInt(delta.AnyLow.Records),
		TimeInExtremeHighRecords:      pointer.FromInt(period.ExtremeHigh.Records),
		TimeInExtremeHighRecordsDelta: pointer.FromInt(delta.ExtremeHigh.Records),
		TimeInHighRecords:             pointer.FromInt(period.High.Records),
		TimeInHighRecordsDelta:        pointer.FromInt(delta.High.Records),
		TimeInLowRecords:              pointer.FromInt(period.Low.Records),
		TimeInLowRecordsDelta:         pointer.FromInt(delta.Low.Records),
		TimeInTargetRecords:           pointer.FromInt(period.Target.Records),
		TimeInTargetRecordsDelta:      pointer.FromInt(delta.Target.Records),
		TimeInVeryHighRecords:         pointer.FromInt(period.VeryHigh.Records),
		TimeInVeryHighRecordsDelta:    pointer.FromInt(delta.VeryHigh.Records),
		TimeInVeryLowRecords:          pointer.FromInt(period.VeryLow.Records),
		TimeInVeryLowRecordsDelta:     pointer.FromInt(delta.VeryLow.Records),
		TotalRecords:                  pointer.FromInt(period.Total.Records),
		TotalRecordsDelta:             pointer.FromInt(delta.Total.Records),
		Min:                           period.Min,
		MinDelta:                      delta.Min,
		Max:                           period.Max,
		MaxDelta:                      delta.Max,
	}

	// reconstruct previous period total records for comparison later
	previousTotalRecords := period.Total.Records - delta.Total.Records

	// percentages should stay nil unless there is records, but schema >5 removed all optional pointers
	if period.Total.Records != 0 {
		exported.HasTimeInTargetPercent = true
		exported.TimeInTargetPercent = pointer.FromFloat64(period.Target.Percent)

		exported.HasTimeInLowPercent = true
		exported.TimeInLowPercent = pointer.FromFloat64(period.Low.Percent)

		exported.HasTimeInVeryLowPercent = true
		exported.TimeInVeryLowPercent = pointer.FromFloat64(period.VeryLow.Percent)

		exported.HasTimeInAnyLowPercent = true
		exported.TimeInAnyLowPercent = pointer.FromFloat64(period.AnyLow.Percent)

		exported.HasTimeInHighPercent = true
		exported.TimeInHighPercent = pointer.FromFloat64(period.High.Percent)

		exported.HasTimeInVeryHighPercent = true
		exported.TimeInVeryHighPercent = pointer.FromFloat64(period.VeryHigh.Percent)

		exported.HasTimeInExtremeHighPercent = true
		exported.TimeInExtremeHighPercent = pointer.FromFloat64(period.ExtremeHigh.Percent)

		exported.HasTimeInAnyHighPercent = true
		exported.TimeInAnyHighPercent = pointer.FromFloat64(period.AnyHigh.Percent)

		exported.HasAverageGlucoseMmol = true
		exported.AverageGlucoseMmol = pointer.FromFloat64(period.AverageGlucose)

		if previousTotalRecords != 0 {
			exported.TimeInTargetPercentDelta = pointer.FromFloat64(delta.Target.Percent)
			exported.TimeInLowPercentDelta = pointer.FromFloat64(delta.Low.Percent)
			exported.TimeInVeryLowPercentDelta = pointer.FromFloat64(delta.VeryLow.Percent)
			exported.TimeInAnyLowPercentDelta = pointer.FromFloat64(delta.AnyLow.Percent)
			exported.TimeInHighPercentDelta = pointer.FromFloat64(delta.High.Percent)
			exported.TimeInVeryHighPercentDelta = pointer.FromFloat64(delta.VeryHigh.Percent)
			exported.TimeInExtremeHighPercentDelta = pointer.FromFloat64(delta.ExtremeHigh.Percent)
			exported.TimeInAnyHighPercentDelta = pointer.FromFloat64(delta.AnyHigh.Percent)
			exported.AverageGlucoseMmolDelta = pointer.FromFloat64(delta.AverageGlucose)
		}
	}

	if period.Total.Records >= 30 && period.DaysWithData >= 7 {
		exported.StandardDeviation = pointer.FromFloat64(period.StandardDeviation)
		exported.StandardDeviationDelta = pointer.FromFloat64(delta.StandardDeviation)
		exported.CoefficientOfVariation = pointer.FromFloat64(period.CoefficientOfVariation)
		exported.CoefficientOfVariationDelta = pointer.FromFloat64(delta.CoefficientOfVariation)
	}

	return exported
}
