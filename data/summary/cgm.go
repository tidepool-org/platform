package summary

//func NewCGMSummary(id string) *CGMSummary {
//	return &CGMSummary{
//		Summary: Summary{
//			UserID: id,
//			Type:   "cgm",
//
//			HasLastUploadDate: false,
//			LastUploadDate:    time.Time{},
//			LastUpdatedDate:   time.Time{},
//			FirstData:         time.Time{},
//			LastData:          nil,
//			OutdatedSince:     nil,
//
//			Config: Config{
//				SchemaVersion:            1,
//				HighGlucoseThreshold:     highBloodGlucose,
//				VeryHighGlucoseThreshold: veryHighBloodGlucose,
//				LowGlucoseThreshold:      lowBloodGlucose,
//				VeryLowGlucoseThreshold:  veryLowBloodGlucose,
//			},
//		},
//		Periods:     make(map[string]*CGMPeriod),
//		HourlyStats: make([]*CGMStats, 0),
//		TotalHours:  0,
//	}
//}
