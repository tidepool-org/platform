package test

import (
	"github.com/tidepool-org/platform/data/types/blood/glucose/summary"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomSummary() *summary.Summary {
	var datum = summary.Summary{
		LastUpdated:              pointer.FromTime(test.RandomTime()),
		FirstData:                pointer.FromTime(test.RandomTime()),
		LastData:                 pointer.FromTime(test.RandomTime()),
		LastUpload:               pointer.FromTime(test.RandomTime()),
		OutdatedSince:            nil,
		TotalDays:                pointer.FromInt(test.RandomIntFromRange(0, 90)),
		HighGlucoseThreshold:     pointer.FromFloat64(test.RandomFloat64FromRange(5, 10)),
		VeryHighGlucoseThreshold: pointer.FromFloat64(test.RandomFloat64FromRange(10, 20)),
		LowGlucoseThreshold:      pointer.FromFloat64(test.RandomFloat64FromRange(3, 5)),
		VeryLowGlucoseThreshold:  pointer.FromFloat64(test.RandomFloat64FromRange(0, 3)),
	}

	// we only make 2, as its lighter and 2 vs 14 vs 90 isn't very different here.
	datum.DailyStats = make([]*summary.Stats, 2)
	for i := 0; i < 2; i++ {
		datum.DailyStats[i] = &summary.Stats{
			DeviceID:        NewDeviceID(),
			Date:            test.RandomTime(),
			TargetMinutes:   test.RandomIntFromRange(0, 1440),
			TargetRecords:   test.RandomIntFromRange(0, 288),
			LowMinutes:      test.RandomIntFromRange(0, 1440),
			LowRecords:      test.RandomIntFromRange(0, 288),
			VeryLowMinutes:  test.RandomIntFromRange(0, 1440),
			VeryLowRecords:  test.RandomIntFromRange(0, 288),
			HighMinutes:     test.RandomIntFromRange(0, 1440),
			HighRecords:     test.RandomIntFromRange(0, 288),
			VeryHighMinutes: test.RandomIntFromRange(0, 1440),
			VeryHighRecords: test.RandomIntFromRange(0, 288),
			TotalGlucose:    test.RandomFloat64FromRange(0, 10000),
			TotalCGMMinutes: test.RandomIntFromRange(0, 1440),
			TotalCGMRecords: test.RandomIntFromRange(0, 288),
			LastRecordTime:  test.RandomTime(),
		}
	}

	datum.Periods = make(map[string]*summary.Period)
	datum.Periods["14d"] = &summary.Period{
		GlucoseManagementIndicator: pointer.FromFloat64(test.RandomFloat64FromRange(0, 20)),
		AverageGlucose: &summary.Glucose{
			Value: pointer.FromFloat64(test.RandomFloat64FromRange(1, 30)),
			Units: pointer.FromString("mmol/L"),
		},
		TimeCGMUsePercent:     pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
		TimeCGMUseMinutes:     pointer.FromInt(test.RandomIntFromRange(0, 129600)),
		TimeCGMUseRecords:     pointer.FromInt(test.RandomIntFromRange(0, 25920)),
		TimeInTargetPercent:   pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
		TimeInTargetMinutes:   pointer.FromInt(test.RandomIntFromRange(0, 129600)),
		TimeInTargetRecords:   pointer.FromInt(test.RandomIntFromRange(0, 25920)),
		TimeInLowPercent:      pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
		TimeInLowMinutes:      pointer.FromInt(test.RandomIntFromRange(0, 129600)),
		TimeInLowRecords:      pointer.FromInt(test.RandomIntFromRange(0, 25920)),
		TimeInVeryLowPercent:  pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
		TimeInVeryLowMinutes:  pointer.FromInt(test.RandomIntFromRange(0, 129600)),
		TimeInVeryLowRecords:  pointer.FromInt(test.RandomIntFromRange(0, 25920)),
		TimeInHighPercent:     pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
		TimeInHighMinutes:     pointer.FromInt(test.RandomIntFromRange(0, 129600)),
		TimeInHighRecords:     pointer.FromInt(test.RandomIntFromRange(0, 25920)),
		TimeInVeryHighPercent: pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
		TimeInVeryHighMinutes: pointer.FromInt(test.RandomIntFromRange(0, 129600)),
		TimeInVeryHighRecords: pointer.FromInt(test.RandomIntFromRange(0, 25920)),
	}

	return &datum
}
