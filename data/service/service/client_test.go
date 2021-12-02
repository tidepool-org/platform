package service_test

import (
	// 	"time"
	// 	"math/rand"
	//
	. "github.com/onsi/ginkgo"
	//. "github.com/onsi/gomega"
	//. "github.com/onsi/gomega/gstruct"
	// 	dataTest "github.com/tidepool-org/platform/data/test"
	// 	userTest "github.com/tidepool-org/platform/user/test"
	// 	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	//
	// 	dataService "github.com/tidepool-org/platform/data/service/service"
)

//
// const unit = "mmol/l"
//
// func NewContinuous(units *string, datumTime *time.Time) *continuous.Continuous {
// 	datum := continuous.New()
// 	datum.Glucose = *dataTypesBloodGlucoseTest.NewGlucose(units)
// 	datum.Type = "cbg"
//
// 	datum.Active = true
// 	datum.ArchivedDataSetID = nil
// 	datum.ArchivedTime = nil
// 	datum.CreatedTime = nil
// 	datum.CreatedUserID = nil
// 	datum.DeletedTime = nil
// 	datum.DeletedUserID = nil
// 	datum.DeviceID = pointer.FromString(deviceID)
// 	datum.ModifiedTime = nil
// 	datum.ModifiedUserID = nil
// 	datum.Time = pointer.FromString(datumTime)
//
// 	return datum
// }
//
// func NewDataSetCGMDataAvg(deviceID string, startTime time.Time, reqAvg float64) []*continuous.Continuous {
// 	dataSetData := []*continuous.Continuous{}
//
// 	// generate 2 weeks of data
// 	for count := 0; count < 4032; count += 2 {
// 		randValue := 1 + (10-1)*rand.Float64()
// 		glucoseValues := [2]float64{randValue+reqAvg, randValue-reqAvg}
//
// 		// this adds 2 entries, one for each side of the average so that the calculated average is the requested value
// 		for i, glucoseValue := range glucoseValues {
// 			datumTime := startTime.Add(time.Duration(-(count+i+1)) * time.Minute * 5).Format(time.RFC3339Nano)
//
// 			datum := NewContinuous(pointer.FromString(unit), &datimTime)
// 			datum.Glucose.Value = pointer.FromFloat64(glucoseValue)
// 			dataSetData = append(dataSetData, datum)
// 		}
// 	}
//
// 	return dataSetData
// }
//
// // creates a dataset with random values evenly divided between ranges
// func NewDataSetCGMDataTimeInRange(deviceID string, startTime time.Time, low float64, high float64) []*continuous.Continuous {
// 	dataSetData := []*continuous.Continuous{}
// 	glucoseBrackets := [3][2]float64{
// 		{1, low},
// 		{low, high},
// 		{high, 20},
// 	}
//
// 	// generate 2 weeks of data
// 	for count := 0; count < 4032; count += 3 {
// 		for i, bracket := range glucoseBrackets {
// 			datumTime := startTime.Add(time.Duration(-(count+i+1)) * time.Minute * 5).Format(time.RFC3339Nano)
//
// 			datum := NewContinuous(pointer.FromString(unit), &datimTime)
// 			datum.Glucose.Value = pointer.FromFloat64(bracket[0] + (bracket[1] - bracket[0]) * rand.Float64())
// 			dataSetData = append(dataSetData, datum)
// 		}
// 	}
//
// 	return dataSetData
// }

var _ = Describe("Client", func() {
	// 	var logger *logTest.Logger
	//
	// 	BeforeEach(func() {
	// 		logger = logTest.NewLogger()
	// 	})
	//
	// 	Context("Summary", func() {
	// 		var ctx context.Context
	// 		var userID string
	// 		var deviceID string
	// 		var dataSetCGM *upload.Upload
	// 		var dataSetCGMData data.Data
	// 		var err error
	//
	// 		BeforeEach(func() {
	// 			ctx = log.NewContextWithLogger(context.Background(), logger)
	// 			userID = userTest.RandomID()
	// 			deviceID = dataTest.NewDeviceID()
	//
	// 			dataSetCGM = NewDataSet(userID, deviceID)
	// 			dataSetCGM.CreatedTime = pointer.FromString("2016-09-01T12:30:00Z")
	//
	// 			dataSetCGMData = NewDataSetCGMDataAvg(deviceID, time.Date(2016, time.Month(9), 1, 12, 30, 0, 0, time.UTC), 5)
	// 		})
	//
	// 		Context("GetDuration", func() {
	// 			var libreDatum continuous.Continuous
	// 			var otherDatum continuous.Continuous
	//
	// 			BeforeEach(func() {
	// 				libreDatum = NewContinuous(pointer.FromString(unit), &datimTime)
	// 				libreDatum.DeviceID = pointer.FromString("a-AbbottFreeStyleLibre-a")
	//
	// 				otherDatum = NewContinuous(pointer.FromString(unit), &datimTime)
	// 			})
	//
	// 			Context("Returns correct 15 minute duration for AbbottFreeStyleLibre", func() {
	// 				duration := dataService.GetDuration(&libreDatum)
	// 				Expect(duration).To(Equal(15))
	// 			})
	//
	// 			Context("Returns correct duration for other devices", func() {
	// 				duration := dataService.GetDuration(&otherDatum)
	// 				Expect(duration).To(Equal(5))
	// 			})
	// 		})
	//
	// 		Context("CalculateWeight", func() {
	// 			Context("Returns correct weight for time range", func() {
	// 				startTime := time.Date(2016, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	// 				endTime := time.Date(2018, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	// 				lastData := time.Date(2017, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	//
	// 				weight, startTime, err := CalculateWeight(startTime, endTime, lastData)
	// 			})
	//
	// 			Context("Returns error on negative time range", func() {
	// 				//CalculateWeight(startTime time.Time, endTime time.Time, lastData time.Time)
	// 			})
	//
	// 			Context("Returns unchanged date and 1 weight when starttime is after lastdata", func() {
	// 				//CalculateWeight(startTime time.Time, endTime time.Time, lastData time.Time)
	// 			})
	// 		})
	//
	// 		Context("CalculateStats", func() {
	// 			Context("Returns correct average glucose for records", func() {
	// 				//CalculateStats(userData []*continuous.Continuous, totalMinutes float64)
	// 			})
	//
	// 			Context("Returns correct time in range value for records", func() {
	// 				//CalculateStats(userData []*continuous.Continuous, totalMinutes float64)
	// 			})
	// 		})
	//
	// 		Context("ReweightStats", func() {
	// 			Context("Returns correctly reweighted stats for 0 weight", func() {
	// 				//ReweightStats(stats *summary.Stats, userSummary *summary.Summary, weight float64)
	// 			})
	//
	// 			Context("Returns correctly reweighted stats for 1 weight", func() {
	// 				//ReweightStats(stats *summary.Stats, userSummary *summary.Summary, weight float64)
	// 			})
	//
	// 			Context("Returns error on negative weight", func() {
	// 				//ReweightStats(stats *summary.Stats, userSummary *summary.Summary, weight float64)
	// 			})
	// 		})
	//
	// 		Context("UpdateSummary", func() {
	// 			//UpdateSummary(ctx context.Context, id string)
	// 		})
	//
	// 		Context("GetAgedSummaries", func() {
	// 			//GetAgedSummaries(ctx context.Context, pagination *page.Pagination)
	// 		})
	//
	// 	})
})
