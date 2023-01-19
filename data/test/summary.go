package test

//
//func RandomSummary() *summary.Summary {
//	var datum = summary.Summary{
//		CGM: summary.CGMSummary{
//			TotalHours:        test.RandomIntFromRange(0, 2160),
//			HasLastUploadDate: test.RandomBool(),
//			LastUploadDate:    test.RandomTime(),
//			LastUpdatedDate:   test.RandomTime(),
//			FirstData:         test.RandomTime(),
//			LastData:          pointer.FromTime(test.RandomTime()),
//			OutdatedSince:     nil,
//		},
//		BGM: summary.BGMSummary{
//			TotalHours:        test.RandomIntFromRange(0, 2160),
//			HasLastUploadDate: test.RandomBool(),
//			LastUploadDate:    test.RandomTime(),
//			LastUpdatedDate:   test.RandomTime(),
//			FirstData:         test.RandomTime(),
//			LastData:          pointer.FromTime(test.RandomTime()),
//			OutdatedSince:     nil,
//		},
//		Config: summary.Config{
//			SchemaVersion:            1,
//			HighGlucoseThreshold:     test.RandomFloat64FromRange(5, 10),
//			VeryHighGlucoseThreshold: test.RandomFloat64FromRange(10, 20),
//			LowGlucoseThreshold:      test.RandomFloat64FromRange(3, 5),
//			VeryLowGlucoseThreshold:  test.RandomFloat64FromRange(0, 3),
//		},
//	}
//
//	// we only make 2, as its lighter and 2 vs 14 vs 90 isn't very different here.
//	datum.CGM.HourlyStats = make([]*summary.CGMStats, 2)
//	for i := 0; i < 2; i++ {
//		datum.CGM.HourlyStats[i] = &summary.CGMStats{
//			Date:            test.RandomTime(),
//			TargetMinutes:   test.RandomIntFromRange(0, 1440),
//			TargetRecords:   test.RandomIntFromRange(0, 288),
//			LowMinutes:      test.RandomIntFromRange(0, 1440),
//			LowRecords:      test.RandomIntFromRange(0, 288),
//			VeryLowMinutes:  test.RandomIntFromRange(0, 1440),
//			VeryLowRecords:  test.RandomIntFromRange(0, 288),
//			HighMinutes:     test.RandomIntFromRange(0, 1440),
//			HighRecords:     test.RandomIntFromRange(0, 288),
//			VeryHighMinutes: test.RandomIntFromRange(0, 1440),
//			VeryHighRecords: test.RandomIntFromRange(0, 288),
//			TotalGlucose:    test.RandomFloat64FromRange(0, 10000),
//			TotalMinutes:    test.RandomIntFromRange(0, 1440),
//			TotalRecords:    test.RandomIntFromRange(0, 288),
//			LastRecordTime:  test.RandomTime(),
//		}
//	}
//
//	// we only make 2, as its lighter and 2 vs 14 vs 90 isn't very different here.
//	datum.BGM.HourlyStats = make([]*summary.BGMStats, 2)
//	for i := 0; i < 2; i++ {
//		datum.BGM.HourlyStats[i] = &summary.BGMStats{
//			Date:            test.RandomTime(),
//			TargetRecords:   test.RandomIntFromRange(0, 288),
//			LowRecords:      test.RandomIntFromRange(0, 288),
//			VeryLowRecords:  test.RandomIntFromRange(0, 288),
//			HighRecords:     test.RandomIntFromRange(0, 288),
//			VeryHighRecords: test.RandomIntFromRange(0, 288),
//			TotalGlucose:    test.RandomFloat64FromRange(0, 10000),
//			TotalRecords:    test.RandomIntFromRange(0, 288),
//			LastRecordTime:  test.RandomTime(),
//		}
//	}
//
//	datum.CGM.Periods = make(map[string]*summary.CGMPeriod)
//	for _, period := range []string{"1d", "7d", "14d", "30d"} {
//		datum.CGM.Periods[period] = &summary.CGMPeriod{
//			GlucoseManagementIndicator:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 20)),
//			HasGlucoseManagementIndicator: test.RandomBool(),
//
//			AverageGlucose: &summary.Glucose{
//				Value: test.RandomFloat64FromRange(1, 30),
//				Units: "mmol/L",
//			},
//			HasAverageGlucose: test.RandomBool(),
//
//			TimeCGMUsePercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
//			HasTimeCGMUsePercent: test.RandomBool(),
//			TimeCGMUseMinutes:    test.RandomIntFromRange(0, 129600),
//			TimeCGMUseRecords:    test.RandomIntFromRange(0, 25920),
//
//			TimeInTargetPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
//			HasTimeInTargetPercent: test.RandomBool(),
//			TimeInTargetMinutes:    test.RandomIntFromRange(0, 129600),
//			TimeInTargetRecords:    test.RandomIntFromRange(0, 25920),
//
//			TimeInLowPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
//			HasTimeInLowPercent: test.RandomBool(),
//			TimeInLowMinutes:    test.RandomIntFromRange(0, 129600),
//			TimeInLowRecords:    test.RandomIntFromRange(0, 25920),
//
//			TimeInVeryLowPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
//			HasTimeInVeryLowPercent: test.RandomBool(),
//			TimeInVeryLowMinutes:    test.RandomIntFromRange(0, 129600),
//			TimeInVeryLowRecords:    test.RandomIntFromRange(0, 25920),
//
//			TimeInHighPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
//			HasTimeInHighPercent: test.RandomBool(),
//			TimeInHighMinutes:    test.RandomIntFromRange(0, 129600),
//			TimeInHighRecords:    test.RandomIntFromRange(0, 25920),
//
//			TimeInVeryHighPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
//			HasTimeInVeryHighPercent: test.RandomBool(),
//			TimeInVeryHighMinutes:    test.RandomIntFromRange(0, 129600),
//			TimeInVeryHighRecords:    test.RandomIntFromRange(0, 25920),
//		}
//	}
//
//	datum.BGM.Periods = make(map[string]*summary.BGMPeriod)
//	for _, period := range []string{"1d", "7d", "14d", "30d"} {
//		datum.BGM.Periods[period] = &summary.BGMPeriod{
//			AverageGlucose: &summary.Glucose{
//				Value: test.RandomFloat64FromRange(1, 30),
//				Units: "mmol/L",
//			},
//			HasAverageGlucose: test.RandomBool(),
//
//			TimeInTargetPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
//			HasTimeInTargetPercent: test.RandomBool(),
//			TimeInTargetRecords:    test.RandomIntFromRange(0, 25920),
//
//			TimeInLowPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
//			HasTimeInLowPercent: test.RandomBool(),
//			TimeInLowRecords:    test.RandomIntFromRange(0, 25920),
//
//			TimeInVeryLowPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
//			HasTimeInVeryLowPercent: test.RandomBool(),
//			TimeInVeryLowRecords:    test.RandomIntFromRange(0, 25920),
//
//			TimeInHighPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
//			HasTimeInHighPercent: test.RandomBool(),
//			TimeInHighRecords:    test.RandomIntFromRange(0, 25920),
//
//			TimeInVeryHighPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
//			HasTimeInVeryHighPercent: test.RandomBool(),
//			TimeInVeryHighRecords:    test.RandomIntFromRange(0, 25920),
//		}
//	}
//
//	return &datum
//}
