package clinics_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/tidepool-org/platform/clinics"
	summaryTest "github.com/tidepool-org/platform/summary/test"
	"github.com/tidepool-org/platform/summary/types"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("NewPatientSummary", func() {
	var userID string
	var cgm *types.CGMSummary
	var bgm *types.BGMSummary

	BeforeEach(func() {
		userID = userTest.RandomUserID()
		cgm = summaryTest.RandomCGMSummary(userID)
		cgm.ID = primitive.NewObjectID()
		bgm = summaryTest.RandomBGMSummary(userID)
		bgm.ID = primitive.NewObjectID()
	})

	It("omits the stats of a summary not given", func() {
		patientSummary := clinics.NewPatientSummary(cgm, nil)
		Expect(patientSummary.CgmStats).ToNot(BeNil())
		Expect(patientSummary.BgmStats).To(BeNil())

		patientSummary = clinics.NewPatientSummary(nil, bgm)
		Expect(patientSummary.CgmStats).To(BeNil())
		Expect(patientSummary.BgmStats).ToNot(BeNil())
	})

	It("reports the id, config, and dates of each summary", func() {
		patientSummary := clinics.NewPatientSummary(cgm, bgm)

		Expect(patientSummary.CgmStats.Id).To(PointTo(Equal(cgm.ID.Hex())))
		Expect(patientSummary.BgmStats.Id).To(PointTo(Equal(bgm.ID.Hex())))

		Expect(patientSummary.CgmStats.Config.SchemaVersion).To(Equal(cgm.Config.SchemaVersion))
		Expect(patientSummary.CgmStats.Config.HighGlucoseThreshold).To(Equal(cgm.Config.HighGlucoseThreshold))
		Expect(patientSummary.CgmStats.Config.VeryHighGlucoseThreshold).To(Equal(cgm.Config.VeryHighGlucoseThreshold))
		Expect(patientSummary.CgmStats.Config.LowGlucoseThreshold).To(Equal(cgm.Config.LowGlucoseThreshold))
		Expect(patientSummary.CgmStats.Config.VeryLowGlucoseThreshold).To(Equal(cgm.Config.VeryLowGlucoseThreshold))

		dates := patientSummary.CgmStats.Dates
		Expect(dates.LastUpdatedDate).To(PointTo(BeTemporally("==", cgm.Dates.LastUpdatedDate)))
		Expect(dates.FirstData).To(PointTo(BeTemporally("==", cgm.Dates.FirstData)))
		Expect(dates.HasFirstData).To(BeTrue())
		Expect(dates.LastData).To(PointTo(BeTemporally("==", cgm.Dates.LastData)))
		Expect(dates.HasLastData).To(BeTrue())
		Expect(dates.LastUploadDate).To(PointTo(BeTemporally("==", cgm.Dates.LastUploadDate)))
		Expect(dates.HasLastUploadDate).To(BeTrue())
		Expect(dates.OutdatedSince).To(Equal(cgm.Dates.OutdatedSince))
		Expect(dates.HasOutdatedSince).To(BeTrue())
		Expect(dates.LastUpdatedReason).To(PointTo(Equal(cgm.Dates.LastUpdatedReason)))
		Expect(dates.OutdatedReason).To(PointTo(Equal(cgm.Dates.OutdatedReason)))
	})

	It("omits the dates the summary does not have", func() {
		cgm.Dates = types.Dates{}

		dates := clinics.NewPatientSummary(cgm, nil).CgmStats.Dates
		Expect(dates.LastUpdatedDate).To(BeNil())
		Expect(dates.FirstData).To(BeNil())
		Expect(dates.HasFirstData).To(BeFalse())
		Expect(dates.LastData).To(BeNil())
		Expect(dates.HasLastData).To(BeFalse())
		Expect(dates.LastUploadDate).To(BeNil())
		Expect(dates.HasLastUploadDate).To(BeFalse())
		Expect(dates.OutdatedSince).To(BeNil())
		Expect(dates.HasOutdatedSince).To(BeFalse())
	})

	// The clinic service replaces only what the report carries, so reasons no longer held must be
	// reported as empty rather than omitted
	It("reports absent reasons as empty", func() {
		cgm.Dates.LastUpdatedReason = nil
		cgm.Dates.OutdatedReason = nil

		dates := clinics.NewPatientSummary(cgm, nil).CgmStats.Dates
		Expect(dates.LastUpdatedReason).To(PointTo(BeEmpty()))
		Expect(dates.OutdatedReason).To(PointTo(BeEmpty()))
	})

	It("drops a period not named as a number of days", func() {
		cgm.Periods.GlucosePeriods["invalid"] = summaryTest.RandomGlucosePeriod(true)

		periods := clinics.NewPatientSummary(cgm, nil).CgmStats.Periods
		Expect(periods).To(HaveLen(4))
		Expect(periods).ToNot(HaveKey("invalid"))
		Expect(periods).To(HaveKey("30d"))
	})

	Context("with a cgm period", func() {
		var period *types.GlucosePeriod

		BeforeEach(func() {
			period = &types.GlucosePeriod{
				GlucoseRanges: types.GlucoseRanges{
					Total:  types.Range{Records: 100, Minutes: 500, Percent: 0.8},
					Target: types.Range{Records: 60, Minutes: 300, Percent: 0.6},
					Low:    types.Range{Records: 10, Minutes: 50, Percent: 0.1},
					High:   types.Range{Records: 30, Minutes: 150, Percent: 0.3},
				},
				MinMax:                     types.MinMax{Min: 2.2, Max: 19.1},
				HoursWithData:              20,
				DaysWithData:               1,
				AverageGlucose:             7.5,
				GlucoseManagementIndicator: 6.9,
				StandardDeviation:          1.2,
				CoefficientOfVariation:     0.3,
				AverageDailyRecords:        288,
				Delta: &types.GlucosePeriod{
					GlucoseRanges: types.GlucoseRanges{
						Total:  types.Range{Records: 40, Minutes: 200, Percent: 0.05},
						Target: types.Range{Records: 20, Minutes: 100, Percent: 0.02},
					},
					AverageGlucose: 0.5,
				},
			}
			cgm.Periods = &types.CGMPeriods{GlucosePeriods: types.GlucosePeriods{"1d": period}}
		})

		It("reports the ranges and their deltas", func() {
			exported := clinics.NewPatientSummary(cgm, nil).CgmStats.Periods["1d"]
			Expect(exported.TotalRecords).To(PointTo(Equal(100)))
			Expect(exported.TotalRecordsDelta).To(PointTo(Equal(40)))
			Expect(exported.TimeCGMUseMinutes).To(PointTo(Equal(500)))
			Expect(exported.TimeCGMUsePercent).To(PointTo(Equal(0.8)))
			Expect(exported.TimeInTargetRecords).To(PointTo(Equal(60)))
			Expect(exported.TimeInTargetRecordsDelta).To(PointTo(Equal(20)))
			Expect(exported.TimeInTargetMinutes).To(PointTo(Equal(300)))
			Expect(exported.TimeInLowRecords).To(PointTo(Equal(10)))
			Expect(exported.TimeInHighRecords).To(PointTo(Equal(30)))
			Expect(exported.TimeInVeryLowRecords).To(PointTo(Equal(0)))
			Expect(exported.HasTotalRecords).To(BeTrue())
			Expect(exported.HasTimeInTargetRecords).To(BeTrue())
			Expect(exported.HasTimeInVeryLowRecords).To(BeFalse())
			Expect(exported.AverageGlucoseMmol).To(PointTo(Equal(7.5)))
			Expect(exported.AverageGlucoseMmolDelta).To(PointTo(Equal(0.5)))
			Expect(exported.StandardDeviation).To(Equal(1.2))
			Expect(exported.CoefficientOfVariation).To(Equal(0.3))
			Expect(exported.AverageDailyRecords).To(PointTo(Equal(288.0)))
			Expect(exported.DaysWithData).To(Equal(1))
			Expect(exported.HoursWithData).To(Equal(20))
			Expect(exported.Min).To(Equal(2.2))
			Expect(exported.Max).To(Equal(19.1))
		})

		It("reports the time-in-range percentages of a single day only above 70 percent use", func() {
			exported := clinics.NewPatientSummary(cgm, nil).CgmStats.Periods["1d"]
			Expect(exported.HasTimeInTargetPercent).To(BeTrue())
			Expect(exported.TimeInTargetPercent).To(PointTo(Equal(0.6)))

			period.Total.Percent = 0.5
			exported = clinics.NewPatientSummary(cgm, nil).CgmStats.Periods["1d"]
			Expect(exported.HasTimeInTargetPercent).To(BeFalse())
			Expect(exported.TimeInTargetPercent).To(BeNil())
		})

		It("reports the time-in-range percentages of a longer period only above a day of use", func() {
			cgm.Periods = &types.CGMPeriods{GlucosePeriods: types.GlucosePeriods{"7d": period}}
			period.Total.Percent = 0.2

			period.Total.Minutes = 1441
			exported := clinics.NewPatientSummary(cgm, nil).CgmStats.Periods["7d"]
			Expect(exported.HasTimeInTargetPercent).To(BeTrue())

			period.Total.Minutes = 1440
			exported = clinics.NewPatientSummary(cgm, nil).CgmStats.Periods["7d"]
			Expect(exported.HasTimeInTargetPercent).To(BeFalse())
		})

		It("reports the glucose management indicator only above 70 percent use", func() {
			exported := clinics.NewPatientSummary(cgm, nil).CgmStats.Periods["1d"]
			Expect(exported.HasGlucoseManagementIndicator).To(BeTrue())
			Expect(exported.GlucoseManagementIndicator).To(PointTo(Equal(6.9)))

			period.Total.Percent = 0.7
			exported = clinics.NewPatientSummary(cgm, nil).CgmStats.Periods["1d"]
			Expect(exported.HasGlucoseManagementIndicator).To(BeFalse())
			Expect(exported.GlucoseManagementIndicator).To(BeNil())
		})

		// The delta reconstructs the previous period, which must fulfill the same requirements
		// before its percentages are compared against
		It("reports the percentage deltas only when the previous period also qualifies", func() {
			// previous use is 0.8 - 0.05 = 0.75
			exported := clinics.NewPatientSummary(cgm, nil).CgmStats.Periods["1d"]
			Expect(exported.TimeInTargetPercentDelta).To(PointTo(Equal(0.02)))
			Expect(exported.GlucoseManagementIndicatorDelta).ToNot(BeNil())

			// previous use is 0.8 - 0.2 = 0.6
			period.Delta.Total.Percent = 0.2
			exported = clinics.NewPatientSummary(cgm, nil).CgmStats.Periods["1d"]
			Expect(exported.TimeInTargetPercentDelta).To(BeNil())
			Expect(exported.GlucoseManagementIndicatorDelta).To(BeNil())
		})

		It("reports no percentages without records", func() {
			period.Total = types.Range{}
			exported := clinics.NewPatientSummary(cgm, nil).CgmStats.Periods["1d"]
			Expect(exported.HasTimeCGMUsePercent).To(BeFalse())
			Expect(exported.TimeCGMUsePercent).To(BeNil())
			Expect(exported.HasAverageGlucoseMmol).To(BeFalse())
			Expect(exported.AverageGlucoseMmol).To(BeNil())
			Expect(exported.HasTimeInTargetPercent).To(BeFalse())
			Expect(exported.HasGlucoseManagementIndicator).To(BeFalse())
		})

		It("reports a period without a delta as changed from nothing", func() {
			period.Delta = nil
			exported := clinics.NewPatientSummary(cgm, nil).CgmStats.Periods["1d"]
			Expect(exported.TotalRecordsDelta).To(PointTo(Equal(0)))
			Expect(exported.TimeInTargetRecordsDelta).To(PointTo(Equal(0)))
		})
	})

	Context("with a bgm period", func() {
		var period *types.GlucosePeriod

		BeforeEach(func() {
			period = &types.GlucosePeriod{
				GlucoseRanges: types.GlucoseRanges{
					Total:  types.Range{Records: 30, Percent: 1},
					Target: types.Range{Records: 20, Percent: 0.67},
				},
				DaysWithData:           7,
				AverageGlucose:         8.1,
				StandardDeviation:      1.4,
				CoefficientOfVariation: 0.4,
				Delta: &types.GlucosePeriod{
					GlucoseRanges: types.GlucoseRanges{
						Total:  types.Range{Records: 10, Percent: 0.1},
						Target: types.Range{Records: 5, Percent: 0.07},
					},
					StandardDeviation:      0.2,
					CoefficientOfVariation: 0.1,
				},
			}
			bgm.Periods = &types.BGMPeriods{GlucosePeriods: types.GlucosePeriods{"30d": period}}
		})

		It("reports the records, percentages, and their deltas", func() {
			exported := clinics.NewPatientSummary(nil, bgm).BgmStats.Periods["30d"]
			Expect(exported.TotalRecords).To(PointTo(Equal(30)))
			Expect(exported.TotalRecordsDelta).To(PointTo(Equal(10)))
			Expect(exported.TimeInTargetRecords).To(PointTo(Equal(20)))
			Expect(exported.TimeInTargetRecordsDelta).To(PointTo(Equal(5)))
			Expect(exported.HasTimeInTargetPercent).To(BeTrue())
			Expect(exported.TimeInTargetPercent).To(PointTo(Equal(0.67)))
			Expect(exported.TimeInTargetPercentDelta).To(PointTo(Equal(0.07)))
			Expect(exported.AverageGlucoseMmol).To(PointTo(Equal(8.1)))
		})

		It("reports no percentages without records", func() {
			period.Total.Records = 0
			exported := clinics.NewPatientSummary(nil, bgm).BgmStats.Periods["30d"]
			Expect(exported.HasTimeInTargetPercent).To(BeFalse())
			Expect(exported.TimeInTargetPercent).To(BeNil())
			Expect(exported.AverageGlucoseMmol).To(BeNil())
		})

		It("reports the deviation measures only with at least 30 records over at least 7 days", func() {
			exported := clinics.NewPatientSummary(nil, bgm).BgmStats.Periods["30d"]
			Expect(exported.StandardDeviation).To(PointTo(Equal(1.4)))
			Expect(exported.StandardDeviationDelta).To(PointTo(Equal(0.2)))
			Expect(exported.CoefficientOfVariation).To(PointTo(Equal(0.4)))
			Expect(exported.CoefficientOfVariationDelta).To(PointTo(Equal(0.1)))

			period.Total.Records = 29
			exported = clinics.NewPatientSummary(nil, bgm).BgmStats.Periods["30d"]
			Expect(exported.StandardDeviation).To(BeNil())
			Expect(exported.CoefficientOfVariation).To(BeNil())

			period.Total.Records = 30
			period.DaysWithData = 6
			exported = clinics.NewPatientSummary(nil, bgm).BgmStats.Periods["30d"]
			Expect(exported.StandardDeviation).To(BeNil())
		})
	})

	It("round-trips a zero-valued outdated since", func() {
		cgm.Dates.OutdatedSince = &time.Time{}
		dates := clinics.NewPatientSummary(cgm, nil).CgmStats.Dates
		Expect(dates.OutdatedSince).To(PointTo(BeTemporally("==", time.Time{})))
		Expect(dates.HasOutdatedSince).To(BeTrue())
	})
})
