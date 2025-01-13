package reporters_test

//		Context("GetPatientsWithRealtimeData", func() {
//
//			It("with some realtime data", func() {
//				endTime := time.Now().UTC().Truncate(time.Hour * 24)
//				startTime := endTime.AddDate(0, 0, -30)
//
//				userContinuousSummary = summaryTest.NewRealtimeSummary(userId, startTime, endTime, 15)
//
//				count := userContinuousSummary.Stats.GetNumberOfDaysWithRealtimeData(startTime, endTime)
//				Expect(count).To(Equal(15))
//			})
//
//			It("with no realtime data", func() {
//				endTime := time.Now().UTC().Truncate(time.Hour * 24)
//				startTime := endTime.AddDate(0, 0, -30)
//
//				userContinuousSummary = summaryTest.NewRealtimeSummary(userId, startTime, endTime, 0)
//
//				count := userContinuousSummary.Stats.GetNumberOfDaysWithRealtimeData(startTime, endTime)
//				Expect(count).To(Equal(0))
//			})
//
//			It("with 60d of realtime data", func() {
//				endTime := time.Now().UTC().Truncate(time.Hour * 24)
//				startTime := endTime.AddDate(0, 0, -30)
//
//				userContinuousSummary = summaryTest.NewRealtimeSummary(userId, startTime, endTime, 60)
//
//				count := userContinuousSummary.Stats.GetNumberOfDaysWithRealtimeData(startTime, endTime)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(count).To(Equal(30))
//			})
//
//			It("with a week of realtime data, with a non-utc, non-dst timezone", func() {
//				loc1 := time.FixedZone("suffering", 12*3600)
//				loc2 := time.FixedZone("pain", 12*3600)
//				lastWeekInNZ := time.Now().In(loc2)
//
//				endTime := time.Date(lastWeekInNZ.Year(), lastWeekInNZ.Month(), lastWeekInNZ.Day(), 23, 59, 59, 0, loc2)
//				startTime := endTime.AddDate(0, 0, -2)
//				startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, loc1)
//
//				userContinuousSummary = summaryTest.NewRealtimeSummary(userId,
//					startTime.AddDate(0, 0, -2),
//					endTime.AddDate(0, 0, 2),
//					7)
//
//				count := userContinuousSummary.Stats.GetNumberOfDaysWithRealtimeData(startTime, endTime)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(count).To(Equal(3))
//			})
//
//			It("with a week of realtime data, with a non-utc, dst timezone", func() {
//				loc1 := time.FixedZone("suffering", 12*3600)
//				loc2 := time.FixedZone("pain", 13*3600)
//				lastWeekInNZ := time.Now().In(loc2)
//
//				endTime := time.Date(lastWeekInNZ.Year(), lastWeekInNZ.Month(), lastWeekInNZ.Day(), 23, 59, 59, 0, loc2)
//				startTime := endTime.AddDate(0, 0, -2)
//				startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, loc1)
//
//				userContinuousSummary = summaryTest.NewRealtimeSummary(userId,
//					startTime.AddDate(0, 0, -2),
//					endTime.AddDate(0, 0, 2),
//					7)
//
//				count := userContinuousSummary.Stats.GetNumberOfDaysWithRealtimeData(startTime, endTime)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(count).To(Equal(3))
//			})
//
//			It("with a week of realtime data, with a non-utc, dst timezone backwards", func() {
//				loc1 := time.FixedZone("pain", 13*3600)
//				loc2 := time.FixedZone("sadness", 12*3600)
//
//				lastWeekInNZ := time.Now().In(loc2)
//
//				endTime := time.Date(lastWeekInNZ.Year(), lastWeekInNZ.Month(), lastWeekInNZ.Day(), 23, 59, 59, 0, loc2)
//				startTime := endTime.AddDate(0, 0, -2)
//				startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, loc1)
//
//				userContinuousSummary = summaryTest.NewRealtimeSummary(userId,
//					startTime.AddDate(0, 0, -2),
//					endTime.AddDate(0, 0, 2),
//					7)
//
//				count := userContinuousSummary.Stats.GetNumberOfDaysWithRealtimeData(startTime, endTime)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(count).To(Equal(3))
//			})
//		})
