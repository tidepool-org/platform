package dexcom_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/dexcom/test"
	"github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("DataRange", func() {
	Describe("Validate", func() {
		DescribeTable("errors when",
			func(setupDataRangeFunc func() *dexcom.DataRange) {
				testDataRange := setupDataRangeFunc()
				validator := validator.New()
				testDataRange.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			},
			Entry("required end is not set", func() *dexcom.DataRange {
				dataRange := test.RandomDataRange()
				dataRange.End = nil
				return dataRange
			}),
			Entry("required start is not set", func() *dexcom.DataRange {
				dataRange := test.RandomDataRange()
				dataRange.Start = nil
				return dataRange
			}),
			Entry("required end display time is not set", func() *dexcom.DataRange {
				dataRange := test.RandomDataRange()
				dataRange.End.DisplayTime = nil
				return dataRange
			}),
			Entry("required start system time is not set", func() *dexcom.DataRange {
				dataRange := test.RandomDataRange()
				dataRange.Start.SystemTime = nil
				return dataRange
			}),
			Entry("required start display time is not set", func() *dexcom.DataRange {
				dataRange := test.RandomDataRange()
				dataRange.Start.DisplayTime = nil
				return dataRange
			}),
		)
	})
})

var _ = Describe("DataRangeResponse", func() {
	Describe("Validate", func() {
		DescribeTable("errors when",
			func(setupDataRangeRespFunc func() *dexcom.DataRangeResponse) {
				testDataRangeResp := setupDataRangeRespFunc()
				validator := validator.New()
				testDataRangeResp.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			},
			Entry("required calibrations are not set", func() *dexcom.DataRangeResponse {
				dataRangeResp := test.RandomDataRangeResponse()
				dataRangeResp.Calibrations = nil
				return dataRangeResp
			}),
			Entry("required events are not set", func() *dexcom.DataRangeResponse {
				dataRangeResp := test.RandomDataRangeResponse()
				dataRangeResp.Events = nil
				return dataRangeResp
			}),
			Entry("required egvs are not set", func() *dexcom.DataRangeResponse {
				dataRangeResp := test.RandomDataRangeResponse()
				dataRangeResp.Egvs = nil
				return dataRangeResp
			}),
		)
	})
	Describe("GetOldestStartDate", func() {

		var oldestTime time.Time
		var nowTime time.Time
		var dayInPast time.Time
		var dayInFuture time.Time

		BeforeEach(func() {
			oldestTime = time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
			nowTime = time.Now().UTC()
			dayInPast = nowTime.Add(-24 * time.Hour)
			dayInFuture = nowTime.Add(24 * time.Hour)
		})

		When("the dates are after now", func() {
			It("it returns now", func() {
				dataRangeResp := test.RandomDataRangeResponseWithDate(dayInFuture)
				Expect(dataRangeResp.GetOldestStartDate().Truncate(5 * time.Minute)).To(Equal(nowTime.Truncate(5 * time.Minute)))
			})
		})
		When("Events start is oldest", func() {
			It("it is returned", func() {
				dataRangeResp := test.RandomDataRangeResponseWithDate(dayInPast)
				dataRangeResp.Events.Start.DisplayTime.Time = oldestTime
				Expect(dataRangeResp.GetOldestStartDate().Equal(oldestTime)).To(BeTrue())
			})
		})
		When("Egvs start is oldest", func() {
			It("it is returned", func() {
				dataRangeResp := test.RandomDataRangeResponseWithDate(dayInPast)
				dataRangeResp.Egvs.Start.DisplayTime.Time = oldestTime
				Expect(dataRangeResp.GetOldestStartDate().Equal(oldestTime)).To(BeTrue())
			})
		})
		When("Calibrations start is oldest", func() {
			It("it is returned", func() {
				dataRangeResp := test.RandomDataRangeResponseWithDate(dayInPast)
				dataRangeResp.Calibrations.Start.DisplayTime.Time = oldestTime
				Expect(dataRangeResp.GetOldestStartDate().Equal(oldestTime)).To(BeTrue())
			})
		})
	})
})
