package test

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomRange(minutes bool) types.Range {
	t := types.Range{
		Glucose:  test.RandomFloat64FromRange(1, 20),
		Percent:  test.RandomFloat64FromRange(0, 1),
		Variance: test.RandomFloat64FromRange(1, 20),
		Records:  test.RandomIntFromRange(1, 12*24*30),
	}

	if minutes {
		t.Minutes = test.RandomIntFromRange(1, 5*12*24*30)
	}

	return t
}

func RandomRanges(minutes bool) types.GlucoseRanges {
	return types.GlucoseRanges{
		Total:       RandomRange(minutes),
		VeryLow:     RandomRange(minutes),
		Low:         RandomRange(minutes),
		Target:      RandomRange(minutes),
		High:        RandomRange(minutes),
		VeryHigh:    RandomRange(minutes),
		ExtremeHigh: RandomRange(minutes),
		AnyLow:      RandomRange(minutes),
		AnyHigh:     RandomRange(minutes),
	}
}

func RandomDates() types.Dates {
	return types.Dates{
		LastUpdatedDate:   test.RandomTime(),
		LastUploadDate:    test.RandomTime(),
		FirstData:         test.RandomTime(),
		LastData:          test.RandomTime(),
		OutdatedSince:     pointer.FromAny(test.RandomTime()),
		OutdatedReason:    []string{"TESTOutdatedReason"},
		LastUpdatedReason: []string{"TESTLastUpdatedReason"},
	}
}

func RandomConfig() types.Config {
	return types.Config{
		SchemaVersion:            test.RandomIntFromRange(1, 5),
		HighGlucoseThreshold:     test.RandomFloat64FromRange(5, 10),
		VeryHighGlucoseThreshold: test.RandomFloat64FromRange(10, 20),
		LowGlucoseThreshold:      test.RandomFloat64FromRange(3, 5),
		VeryLowGlucoseThreshold:  test.RandomFloat64FromRange(0, 3),
	}
}

func RandomGlucosePeriod(minutes bool) *types.GlucosePeriod {
	return &types.GlucosePeriod{
		GlucoseRanges:              RandomRanges(minutes),
		HoursWithData:              test.RandomIntFromRange(1, 1440),
		DaysWithData:               test.RandomIntFromRange(1, 30),
		AverageGlucose:             test.RandomFloat64FromRange(1, 20),
		GlucoseManagementIndicator: test.RandomFloat64FromRange(1, 20),
		CoefficientOfVariation:     test.RandomFloat64FromRange(1, 20),
		StandardDeviation:          test.RandomFloat64FromRange(1, 20),
		AverageDailyRecords:        test.RandomFloat64FromRange(1, 288),
		Delta:                      nil,
	}
}

func RandomContinuousPeriod() *types.ContinuousPeriod {
	return &types.ContinuousPeriod{
		ContinuousRanges: types.ContinuousRanges{
			Realtime: RandomRange(true),
			Deferred: RandomRange(true),
			Total:    RandomRange(true),
		},
		AverageDailyRecords: test.RandomFloat64FromRange(1, 288),
	}
}

func RandomCGMSummary(userId string) *types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket] {
	datum := types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]{
		BaseSummary: types.BaseSummary{
			Type:   "cgm",
			UserID: userId,
			Config: RandomConfig(),
			Dates:  RandomDates(),
		},
		Periods: &types.CGMPeriods{},
	}

	for _, period := range []string{"1d", "7d", "14d", "30d"} {
		datum.Periods.GlucosePeriods[period] = RandomGlucosePeriod(true)
		datum.Periods.GlucosePeriods[period].Delta = RandomGlucosePeriod(true)
	}

	return &datum
}

func RandomBGMSummary(userId string) *types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket] {
	datum := types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
		BaseSummary: types.BaseSummary{
			Type:   "bgm",
			UserID: userId,
			Config: RandomConfig(),
			Dates:  RandomDates(),
		},
		Periods: &types.BGMPeriods{},
	}

	for _, period := range []string{"1d", "7d", "14d", "30d"} {
		datum.Periods.GlucosePeriods[period] = RandomGlucosePeriod(false)
		datum.Periods.GlucosePeriods[period].Delta = RandomGlucosePeriod(false)
	}

	return &datum
}

func RandomContinuousSummary(userId string) *types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket] {
	datum := types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
		BaseSummary: types.BaseSummary{
			Type:   "con",
			UserID: userId,
			Config: RandomConfig(),
			Dates:  RandomDates(),
		},
		Periods: &types.ContinuousPeriods{},
	}

	for _, period := range []string{"30d"} {
		(*datum.Periods)[period] = RandomContinuousPeriod()
	}

	return &datum
}

func NewRealtimeBuckets(userId string, startTime time.Time, endTime time.Time, realtimeDays int) []*types.Bucket[*types.ContinuousBucket, types.ContinuousBucket] {
	totalHours := int(endTime.Sub(startTime).Hours())

	buckets := make([]*types.Bucket[*types.ContinuousBucket, types.ContinuousBucket], totalHours)

	var yesterday time.Time
	var today time.Time
	var bucketDate time.Time
	var flaggedDays int
	var recordCount int

	for i := 0; i < len(buckets); i++ {
		bucketDate = startTime.Add(time.Duration(i) * time.Hour)
		today = bucketDate.Truncate(time.Hour * 24)

		if flaggedDays < realtimeDays {
			recordCount = test.RandomIntFromRange(1, 12)

			if today.After(yesterday) {
				flaggedDays++
				yesterday = today
			}

		} else {
			recordCount = 0
		}

		buckets[i] = &types.Bucket[*types.ContinuousBucket, types.ContinuousBucket]{
			BaseBucket: types.BaseBucket{
				ID:        primitive.ObjectID{},
				UserId:    userId,
				Type:      types.SummaryTypeContinuous,
				Time:      bucketDate,
				FirstData: time.Time{},
				LastData:  time.Time{},
			},
			Data: &types.ContinuousBucket{
				ContinuousRanges: types.ContinuousRanges{
					Realtime: types.Range{Records: recordCount},
					Deferred: types.Range{Records: recordCount},
				},
			},
		}
	}

	return buckets
}
