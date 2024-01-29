package dexcom_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
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
				dataRangeResp.EGVs = nil
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
			It("it returns an error", func() {
				dataRangeResp := test.RandomDataRangeResponseWithDate(dayInFuture)
				oldest, err := dataRangeResp.GetOldestStartDate()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("the oldest start date should before now"))
				Expect(oldest.IsZero()).To(Equal(true))
			})
		})
		When("the dates is zero", func() {
			It("it returns an error", func() {
				dataRangeResp := test.RandomDataRangeResponseWithDate(dayInPast)
				dataRangeResp.Events.Start.DisplayTime.Time = time.Time{}
				oldest, err := dataRangeResp.GetOldestStartDate()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("invalid start display time"))
				Expect(oldest.IsZero()).To(Equal(true))
			})
		})
		When("start not set", func() {
			It("it returns an error", func() {
				dataRangeResp := test.RandomDataRangeResponseWithDate(dayInPast)
				dataRangeResp.EGVs.Start = nil
				oldest, err := dataRangeResp.GetOldestStartDate()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("invalid start display time"))
				Expect(oldest.IsZero()).To(Equal(true))
			})
		})
		When("Events start is oldest", func() {
			It("it is returned", func() {
				dataRangeResp := test.RandomDataRangeResponseWithDate(dayInPast)
				dataRangeResp.Events.Start.DisplayTime.Time = oldestTime
				oldest, err := dataRangeResp.GetOldestStartDate()
				Expect(err).To(BeNil())
				Expect(oldest.Equal(oldestTime)).To(BeTrue())
			})
		})
		When("Egvs start is oldest", func() {
			It("it is returned", func() {
				dataRangeResp := test.RandomDataRangeResponseWithDate(dayInPast)
				dataRangeResp.EGVs.Start.DisplayTime.Time = oldestTime
				oldest, err := dataRangeResp.GetOldestStartDate()
				Expect(err).To(BeNil())
				Expect(oldest.Equal(oldestTime)).To(BeTrue())
			})
		})
		When("Calibrations start is oldest", func() {
			It("it is returned", func() {
				dataRangeResp := test.RandomDataRangeResponseWithDate(dayInPast)
				dataRangeResp.Calibrations.Start.DisplayTime.Time = oldestTime
				oldest, err := dataRangeResp.GetOldestStartDate()
				Expect(err).To(BeNil())
				Expect(oldest.Equal(oldestTime)).To(BeTrue())
			})
		})
	})
})
