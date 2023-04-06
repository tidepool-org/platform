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
			HasLastUploadDate: test.RandomBool(),
			LastUploadDate:    test.RandomTime(),
			LastUpdatedDate:   test.RandomTime(),
			FirstData:         test.RandomTime(),
			LastData:          pointer.FromAny(test.RandomTime()),
			OutdatedSince:     pointer.FromAny(test.RandomTime()),
		},
		Stats: &types.CGMStats{
			TotalHours: test.RandomIntFromRange(1, 720),
			Periods:    make(map[string]types.CGMPeriod),

			// we only make 2, as its lighter and 2 vs 14 vs 90 isn't very different here.
			Buckets: make(types.Buckets[types.CGMBucketData, *types.CGMBucketData], 2),
		},
	}

	for i := 0; i < len(datum.Stats.Buckets); i++ {
		datum.Stats.Buckets[i] = types.Bucket[types.CGMBucketData, *types.CGMBucketData]{
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
		datum.Stats.Periods[period] = types.CGMPeriod{
			GlucoseManagementIndicator:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 20)),
			HasGlucoseManagementIndicator: test.RandomBool(),

			AverageGlucose: &types.Glucose{
				Value: test.RandomFloat64FromRange(1, 30),
				Units: "mmol/L",
			},
			HasAverageGlucose: test.RandomBool(),

			TimeCGMUsePercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
			HasTimeCGMUsePercent: test.RandomBool(),
			TimeCGMUseMinutes:    test.RandomIntFromRange(0, 129600),
			TimeCGMUseRecords:    test.RandomIntFromRange(0, 25920),

			TimeInTargetPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
			HasTimeInTargetPercent: test.RandomBool(),
			TimeInTargetMinutes:    test.RandomIntFromRange(0, 129600),
			TimeInTargetRecords:    test.RandomIntFromRange(0, 25920),

			TimeInLowPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
			HasTimeInLowPercent: test.RandomBool(),
			TimeInLowMinutes:    test.RandomIntFromRange(0, 129600),
			TimeInLowRecords:    test.RandomIntFromRange(0, 25920),

			TimeInVeryLowPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
			HasTimeInVeryLowPercent: test.RandomBool(),
			TimeInVeryLowMinutes:    test.RandomIntFromRange(0, 129600),
			TimeInVeryLowRecords:    test.RandomIntFromRange(0, 25920),

			TimeInHighPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
			HasTimeInHighPercent: test.RandomBool(),
			TimeInHighMinutes:    test.RandomIntFromRange(0, 129600),
			TimeInHighRecords:    test.RandomIntFromRange(0, 25920),

			TimeInVeryHighPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
			HasTimeInVeryHighPercent: test.RandomBool(),
			TimeInVeryHighMinutes:    test.RandomIntFromRange(0, 129600),
			TimeInVeryHighRecords:    test.RandomIntFromRange(0, 25920),
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
			HasLastUploadDate: test.RandomBool(),
			LastUploadDate:    test.RandomTime(),
			LastUpdatedDate:   test.RandomTime(),
			FirstData:         test.RandomTime(),
			LastData:          pointer.FromAny(test.RandomTime()),
			OutdatedSince:     pointer.FromAny(test.RandomTime()),
		},
		Stats: &types.BGMStats{
			TotalHours: test.RandomIntFromRange(1, 720),
			Periods:    make(map[string]types.BGMPeriod),

			// we only make 2, as its lighter and 2 vs 14 vs 90 isn't very different here.
			Buckets: make(types.Buckets[types.BGMBucketData, *types.BGMBucketData], 2),
		},
	}

	for i := 0; i < len(datum.Stats.Buckets); i++ {
		datum.Stats.Buckets[i] = types.Bucket[types.BGMBucketData, *types.BGMBucketData]{
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
		datum.Stats.Periods[period] = types.BGMPeriod{
			AverageGlucose: &types.Glucose{
				Value: test.RandomFloat64FromRange(1, 30),
				Units: "mmol/L",
			},
			HasAverageGlucose: test.RandomBool(),

			TimeInTargetPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
			HasTimeInTargetPercent: test.RandomBool(),
			TimeInTargetRecords:    test.RandomIntFromRange(0, 25920),

			TimeInLowPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
			HasTimeInLowPercent: test.RandomBool(),
			TimeInLowRecords:    test.RandomIntFromRange(0, 25920),

			TimeInVeryLowPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
			HasTimeInVeryLowPercent: test.RandomBool(),
			TimeInVeryLowRecords:    test.RandomIntFromRange(0, 25920),

			TimeInHighPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
			HasTimeInHighPercent: test.RandomBool(),
			TimeInHighRecords:    test.RandomIntFromRange(0, 25920),

			TimeInVeryHighPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
			HasTimeInVeryHighPercent: test.RandomBool(),
			TimeInVeryHighRecords:    test.RandomIntFromRange(0, 25920),
		}
	}

	return &datum
}
