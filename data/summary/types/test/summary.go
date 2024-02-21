package test

import (
	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomCGMSummary(userId string) *types.Summary[types.CGMStats, *types.CGMStats] {
	datum := types.Summary[types.CGMStats, *types.CGMStats]{
		UserID: userId,
		Type:   "cgm",
		Config: types.Config{
			SchemaVersion:            test.RandomIntFromRange(1, 5),
			HighGlucoseThreshold:     test.RandomFloat64FromRange(5, 10),
			VeryHighGlucoseThreshold: test.RandomFloat64FromRange(10, 20),
			LowGlucoseThreshold:      test.RandomFloat64FromRange(3, 5),
			VeryLowGlucoseThreshold:  test.RandomFloat64FromRange(0, 3),
		},
		Dates: types.Dates{
			LastUpdatedDate:   test.RandomTime(),
			HasLastUploadDate: test.RandomBool(),
			LastUploadDate:    pointer.FromAny(test.RandomTime()),
			HasFirstData:      test.RandomBool(),
			FirstData:         pointer.FromAny(test.RandomTime()),
			HasLastData:       test.RandomBool(),
			LastData:          pointer.FromAny(test.RandomTime()),
			HasOutdatedSince:  test.RandomBool(),
			OutdatedSince:     pointer.FromAny(test.RandomTime()),
		},
		Stats: &types.CGMStats{
			TotalHours:    test.RandomIntFromRange(1, 720),
			Periods:       make(map[string]*types.CGMPeriod),
			OffsetPeriods: make(map[string]*types.CGMPeriod),

			// we only make 2, as its lighter and 2 vs 14 vs 90 isn't very different here.
			Buckets: make([]*types.Bucket[*types.CGMBucketData, types.CGMBucketData], 2),
		},
	}

	for i := 0; i < len(datum.Stats.Buckets); i++ {
		datum.Stats.Buckets[i] = &types.Bucket[*types.CGMBucketData, types.CGMBucketData]{
			Date:           test.RandomTime(),
			LastRecordTime: test.RandomTime(),
			Data: &types.CGMBucketData{
				LastRecordDuration: test.RandomIntFromRange(1, 10),
				TargetMinutes:      test.RandomIntFromRange(0, 1440),
				TargetRecords:      test.RandomIntFromRange(0, 288),
				LowMinutes:         test.RandomIntFromRange(0, 1440),
				LowRecords:         test.RandomIntFromRange(0, 288),
				VeryLowMinutes:     test.RandomIntFromRange(0, 1440),
				VeryLowRecords:     test.RandomIntFromRange(0, 288),
				HighMinutes:        test.RandomIntFromRange(0, 1440),
				HighRecords:        test.RandomIntFromRange(0, 288),
				VeryHighMinutes:    test.RandomIntFromRange(0, 1440),
				VeryHighRecords:    test.RandomIntFromRange(0, 288),
				TotalGlucose:       test.RandomFloat64FromRange(0, 10000),
				TotalMinutes:       test.RandomIntFromRange(0, 1440),
				TotalRecords:       test.RandomIntFromRange(0, 288),
			},
		}
	}

	for _, period := range []string{"1d", "7d", "14d", "30d"} {
		datum.Stats.Periods[period] = &types.CGMPeriod{
			HasGlucoseManagementIndicator:   test.RandomBool(),
			GlucoseManagementIndicator:      pointer.FromAny(test.RandomFloat64FromRange(0, 20)),
			GlucoseManagementIndicatorDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 20)),

			HasAverageGlucoseMmol:   test.RandomBool(),
			AverageGlucoseMmol:      pointer.FromAny(test.RandomFloat64FromRange(1, 30)),
			AverageGlucoseMmolDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 20)),

			HasTotalRecords:   test.RandomBool(),
			TotalRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TotalRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasAverageDailyRecords:   test.RandomBool(),
			AverageDailyRecords:      pointer.FromAny(test.RandomFloat64FromRange(0, 25920)),
			AverageDailyRecordsDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 25920)),

			HasTimeCGMUsePercent:   test.RandomBool(),
			TimeCGMUsePercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeCGMUsePercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeCGMUseMinutes:   test.RandomBool(),
			TimeCGMUseMinutes:      pointer.FromAny(test.RandomIntFromRange(0, 129600)),
			TimeCGMUseMinutesDelta: pointer.FromAny(test.RandomIntFromRange(0, 129600)),

			HasTimeCGMUseRecords:   test.RandomBool(),
			TimeCGMUseRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeCGMUseRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasTimeInTargetPercent:   test.RandomBool(),
			TimeInTargetPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInTargetPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInTargetMinutes:   test.RandomBool(),
			TimeInTargetMinutes:      pointer.FromAny(test.RandomIntFromRange(0, 129600)),
			TimeInTargetMinutesDelta: pointer.FromAny(test.RandomIntFromRange(0, 129600)),

			HasTimeInTargetRecords:   test.RandomBool(),
			TimeInTargetRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInTargetRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasTimeInLowPercent:   test.RandomBool(),
			TimeInLowPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInLowPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInLowMinutes:   test.RandomBool(),
			TimeInLowMinutes:      pointer.FromAny(test.RandomIntFromRange(0, 129600)),
			TimeInLowMinutesDelta: pointer.FromAny(test.RandomIntFromRange(0, 129600)),

			HasTimeInLowRecords:   test.RandomBool(),
			TimeInLowRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInLowRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasTimeInVeryLowPercent:   test.RandomBool(),
			TimeInVeryLowPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInVeryLowPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInVeryLowMinutes:   test.RandomBool(),
			TimeInVeryLowMinutes:      pointer.FromAny(test.RandomIntFromRange(0, 129600)),
			TimeInVeryLowMinutesDelta: pointer.FromAny(test.RandomIntFromRange(0, 129600)),

			HasTimeInVeryLowRecords:   test.RandomBool(),
			TimeInVeryLowRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInVeryLowRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasTimeInHighPercent:   test.RandomBool(),
			TimeInHighPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInHighPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInHighMinutes:   test.RandomBool(),
			TimeInHighMinutes:      pointer.FromAny(test.RandomIntFromRange(0, 129600)),
			TimeInHighMinutesDelta: pointer.FromAny(test.RandomIntFromRange(0, 129600)),

			HasTimeInHighRecords:   test.RandomBool(),
			TimeInHighRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInHighRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasTimeInVeryHighPercent:   test.RandomBool(),
			TimeInVeryHighPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInVeryHighPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInVeryHighMinutes:   test.RandomBool(),
			TimeInVeryHighMinutes:      pointer.FromAny(test.RandomIntFromRange(0, 129600)),
			TimeInVeryHighMinutesDelta: pointer.FromAny(test.RandomIntFromRange(0, 129600)),

			HasTimeInVeryHighRecords:   test.RandomBool(),
			TimeInVeryHighRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInVeryHighRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),
		}

		datum.Stats.OffsetPeriods[period] = &types.CGMPeriod{
			HasGlucoseManagementIndicator:   test.RandomBool(),
			GlucoseManagementIndicator:      pointer.FromAny(test.RandomFloat64FromRange(0, 20)),
			GlucoseManagementIndicatorDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 20)),

			HasAverageGlucoseMmol:   test.RandomBool(),
			AverageGlucoseMmol:      pointer.FromAny(test.RandomFloat64FromRange(1, 30)),
			AverageGlucoseMmolDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 20)),

			HasTotalRecords:   test.RandomBool(),
			TotalRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TotalRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasAverageDailyRecords:   test.RandomBool(),
			AverageDailyRecords:      pointer.FromAny(test.RandomFloat64FromRange(0, 25920)),
			AverageDailyRecordsDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 25920)),

			HasTimeCGMUsePercent:   test.RandomBool(),
			TimeCGMUsePercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeCGMUsePercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeCGMUseMinutes:   test.RandomBool(),
			TimeCGMUseMinutes:      pointer.FromAny(test.RandomIntFromRange(0, 129600)),
			TimeCGMUseMinutesDelta: pointer.FromAny(test.RandomIntFromRange(0, 129600)),

			HasTimeCGMUseRecords:   test.RandomBool(),
			TimeCGMUseRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeCGMUseRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasTimeInTargetPercent:   test.RandomBool(),
			TimeInTargetPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInTargetPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInTargetMinutes:   test.RandomBool(),
			TimeInTargetMinutes:      pointer.FromAny(test.RandomIntFromRange(0, 129600)),
			TimeInTargetMinutesDelta: pointer.FromAny(test.RandomIntFromRange(0, 129600)),

			HasTimeInTargetRecords:   test.RandomBool(),
			TimeInTargetRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInTargetRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasTimeInLowPercent:   test.RandomBool(),
			TimeInLowPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInLowPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInLowMinutes:   test.RandomBool(),
			TimeInLowMinutes:      pointer.FromAny(test.RandomIntFromRange(0, 129600)),
			TimeInLowMinutesDelta: pointer.FromAny(test.RandomIntFromRange(0, 129600)),

			HasTimeInLowRecords:   test.RandomBool(),
			TimeInLowRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInLowRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasTimeInVeryLowPercent:   test.RandomBool(),
			TimeInVeryLowPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInVeryLowPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInVeryLowMinutes:   test.RandomBool(),
			TimeInVeryLowMinutes:      pointer.FromAny(test.RandomIntFromRange(0, 129600)),
			TimeInVeryLowMinutesDelta: pointer.FromAny(test.RandomIntFromRange(0, 129600)),

			HasTimeInVeryLowRecords:   test.RandomBool(),
			TimeInVeryLowRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInVeryLowRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasTimeInHighPercent:   test.RandomBool(),
			TimeInHighPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInHighPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInHighMinutes:   test.RandomBool(),
			TimeInHighMinutes:      pointer.FromAny(test.RandomIntFromRange(0, 129600)),
			TimeInHighMinutesDelta: pointer.FromAny(test.RandomIntFromRange(0, 129600)),

			HasTimeInHighRecords:   test.RandomBool(),
			TimeInHighRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInHighRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasTimeInVeryHighPercent:   test.RandomBool(),
			TimeInVeryHighPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInVeryHighPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInVeryHighMinutes:   test.RandomBool(),
			TimeInVeryHighMinutes:      pointer.FromAny(test.RandomIntFromRange(0, 129600)),
			TimeInVeryHighMinutesDelta: pointer.FromAny(test.RandomIntFromRange(0, 129600)),

			HasTimeInVeryHighRecords:   test.RandomBool(),
			TimeInVeryHighRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInVeryHighRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),
		}
	}

	return &datum
}

func RandomBGMSummary(userId string) *types.Summary[types.BGMStats, *types.BGMStats] {
	datum := types.Summary[types.BGMStats, *types.BGMStats]{
		UserID: userId,
		Type:   "bgm",
		Config: types.Config{
			SchemaVersion:            test.RandomIntFromRange(1, 5),
			HighGlucoseThreshold:     test.RandomFloat64FromRange(5, 10),
			VeryHighGlucoseThreshold: test.RandomFloat64FromRange(10, 20),
			LowGlucoseThreshold:      test.RandomFloat64FromRange(3, 5),
			VeryLowGlucoseThreshold:  test.RandomFloat64FromRange(0, 3),
		},
		Dates: types.Dates{
			LastUpdatedDate:   test.RandomTime(),
			HasLastUploadDate: test.RandomBool(),
			LastUploadDate:    pointer.FromAny(test.RandomTime()),
			HasFirstData:      test.RandomBool(),
			FirstData:         pointer.FromAny(test.RandomTime()),
			HasLastData:       test.RandomBool(),
			LastData:          pointer.FromAny(test.RandomTime()),
			HasOutdatedSince:  test.RandomBool(),
			OutdatedSince:     pointer.FromAny(test.RandomTime()),
		},
		Stats: &types.BGMStats{
			TotalHours:    test.RandomIntFromRange(1, 720),
			Periods:       make(map[string]*types.BGMPeriod),
			OffsetPeriods: make(map[string]*types.BGMPeriod),

			// we only make 2, as its lighter and 2 vs 14 vs 90 isn't very different here.
			Buckets: make([]*types.Bucket[*types.BGMBucketData, types.BGMBucketData], 2),
		},
	}

	for i := 0; i < len(datum.Stats.Buckets); i++ {
		datum.Stats.Buckets[i] = &types.Bucket[*types.BGMBucketData, types.BGMBucketData]{
			Date:           test.RandomTime(),
			LastRecordTime: test.RandomTime(),
			Data: &types.BGMBucketData{
				TargetRecords:   test.RandomIntFromRange(0, 288),
				LowRecords:      test.RandomIntFromRange(0, 288),
				VeryLowRecords:  test.RandomIntFromRange(0, 288),
				HighRecords:     test.RandomIntFromRange(0, 288),
				VeryHighRecords: test.RandomIntFromRange(0, 288),
				TotalGlucose:    test.RandomFloat64FromRange(0, 10000),
				TotalRecords:    test.RandomIntFromRange(0, 288),
			},
		}
	}

	for _, period := range []string{"1d", "7d", "14d", "30d"} {
		datum.Stats.Periods[period] = &types.BGMPeriod{
			HasAverageGlucoseMmol:   test.RandomBool(),
			AverageGlucoseMmol:      pointer.FromAny(test.RandomFloat64FromRange(1, 30)),
			AverageGlucoseMmolDelta: pointer.FromAny(test.RandomFloat64FromRange(1, 30)),

			HasTotalRecords:   test.RandomBool(),
			TotalRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TotalRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasAverageDailyRecords:   test.RandomBool(),
			AverageDailyRecords:      pointer.FromAny(test.RandomFloat64FromRange(0, 25920)),
			AverageDailyRecordsDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 25920)),

			HasTimeInTargetPercent:   test.RandomBool(),
			TimeInTargetPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInTargetPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInTargetRecords:   test.RandomBool(),
			TimeInTargetRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInTargetRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasTimeInLowPercent:   test.RandomBool(),
			TimeInLowPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInLowPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInLowRecords:   test.RandomBool(),
			TimeInLowRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInLowRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasTimeInVeryLowPercent:   test.RandomBool(),
			TimeInVeryLowPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInVeryLowPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInVeryLowRecords:   test.RandomBool(),
			TimeInVeryLowRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInVeryLowRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasTimeInHighPercent:   test.RandomBool(),
			TimeInHighPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInHighPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInHighRecords:   test.RandomBool(),
			TimeInHighRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInHighRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasTimeInVeryHighPercent:   test.RandomBool(),
			TimeInVeryHighPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInVeryHighPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInVeryHighRecords:   test.RandomBool(),
			TimeInVeryHighRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInVeryHighRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),
		}

		datum.Stats.OffsetPeriods[period] = &types.BGMPeriod{
			HasAverageGlucoseMmol:   test.RandomBool(),
			AverageGlucoseMmol:      pointer.FromAny(test.RandomFloat64FromRange(1, 30)),
			AverageGlucoseMmolDelta: pointer.FromAny(test.RandomFloat64FromRange(1, 30)),

			HasTotalRecords:   test.RandomBool(),
			TotalRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TotalRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasAverageDailyRecords:   test.RandomBool(),
			AverageDailyRecords:      pointer.FromAny(test.RandomFloat64FromRange(0, 25920)),
			AverageDailyRecordsDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 25920)),

			HasTimeInTargetPercent:   test.RandomBool(),
			TimeInTargetPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInTargetPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInTargetRecords:   test.RandomBool(),
			TimeInTargetRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInTargetRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasTimeInLowPercent:   test.RandomBool(),
			TimeInLowPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInLowPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInLowRecords:   test.RandomBool(),
			TimeInLowRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInLowRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasTimeInVeryLowPercent:   test.RandomBool(),
			TimeInVeryLowPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInVeryLowPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInVeryLowRecords:   test.RandomBool(),
			TimeInVeryLowRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInVeryLowRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasTimeInHighPercent:   test.RandomBool(),
			TimeInHighPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInHighPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInHighRecords:   test.RandomBool(),
			TimeInHighRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInHighRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),

			HasTimeInVeryHighPercent:   test.RandomBool(),
			TimeInVeryHighPercent:      pointer.FromAny(test.RandomFloat64FromRange(0, 1)),
			TimeInVeryHighPercentDelta: pointer.FromAny(test.RandomFloat64FromRange(0, 1)),

			HasTimeInVeryHighRecords:   test.RandomBool(),
			TimeInVeryHighRecords:      pointer.FromAny(test.RandomIntFromRange(0, 25920)),
			TimeInVeryHighRecordsDelta: pointer.FromAny(test.RandomIntFromRange(0, 25920)),
		}
	}

	return &datum
}
