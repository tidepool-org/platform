package types_test

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	. "github.com/tidepool-org/platform/summary/test"
	. "github.com/tidepool-org/platform/summary/types"
)

var _ = Describe("Glucose", func() {
	var bucketTime time.Time
	var err error
	var userId string
	var now time.Time

	BeforeEach(func() {
		now = time.Now()
		userId = "1234"
		bucketTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	})

	Context("MinMax", func() {
		Context("Update", func() {

			It("Adding 1 value", func() {
				minMax := MinMax{}
				datum := NewGlucoseWithValue(continuous.Type, now, 5)

				minMax.Update(datum)
				Expect(minMax.Max).To(Equal(5.0))
				Expect(minMax.Min).To(Equal(5.0))
			})

			It("Adding 2 values, replacing Max", func() {
				minMax := MinMax{}

				By("adding the first number")
				datum := NewGlucoseWithValue(continuous.Type, now, 3)
				minMax.Update(datum)
				Expect(minMax.Max).To(Equal(3.0))
				Expect(minMax.Min).To(Equal(3.0))

				By("adding the second number")
				datum = NewGlucoseWithValue(continuous.Type, now, 7)
				minMax.Update(datum)
				Expect(minMax.Max).To(Equal(7.0))
				Expect(minMax.Min).To(Equal(3.0))
			})

			It("Adding 2 values, replacing Min", func() {
				minMax := MinMax{}

				By("adding the first number")
				datum := NewGlucoseWithValue(continuous.Type, now, 7)
				minMax.Update(datum)
				Expect(minMax.Max).To(Equal(7.0))
				Expect(minMax.Min).To(Equal(7.0))

				By("adding the second number")
				datum = NewGlucoseWithValue(continuous.Type, now, 3)
				minMax.Update(datum)
				Expect(minMax.Max).To(Equal(7.0))
				Expect(minMax.Min).To(Equal(3.0))
			})

			It("Adding one value, leaving unchanged", func() {
				minMax := MinMax{Min: 3, Max: 7}

				datum := NewGlucoseWithValue(continuous.Type, now, 5)
				minMax.Update(datum)
				Expect(minMax.Max).To(Equal(7.0))
				Expect(minMax.Min).To(Equal(3.0))
			})

		})

		Context("Add", func() {

			It("replacing Min", func() {
				minMax := MinMax{Min: 3, Max: 7}
				minMaxNew := MinMax{Min: 2, Max: 7}

				minMax.Add(&minMaxNew)
				Expect(minMax.Max).To(Equal(7.0))
				Expect(minMax.Min).To(Equal(2.0))
			})

			It("replacing Max", func() {
				minMax := MinMax{Min: 3, Max: 4}
				minMaxNew := MinMax{Min: 3, Max: 7}

				minMax.Add(&minMaxNew)
				Expect(minMax.Max).To(Equal(7.0))
				Expect(minMax.Min).To(Equal(3.0))
			})

			It("replacing both", func() {
				minMax := MinMax{Min: 4, Max: 5}
				minMaxNew := MinMax{Min: 3, Max: 7}

				minMax.Add(&minMaxNew)
				Expect(minMax.Max).To(Equal(7.0))
				Expect(minMax.Min).To(Equal(3.0))
			})

			It("replacing neither", func() {
				minMax := MinMax{Min: 3, Max: 7}
				minMaxNew := MinMax{Min: 4, Max: 5}

				minMax.Add(&minMaxNew)
				Expect(minMax.Max).To(Equal(7.0))
				Expect(minMax.Min).To(Equal(3.0))
			})

		})
	})

	Context("Range", func() {
		It("range.UpdateTotal", func() {
			glucoseRange := Range{}
			datum := NewGlucoseWithValue(continuous.Type, now, 5)

			By("adding 5 minutes of 5mmol")
			glucoseRange.UpdateTotal(datum)
			Expect(glucoseRange.Glucose).To(Equal(5.0 * 5.0))
			Expect(glucoseRange.Records).To(Equal(1))
			Expect(glucoseRange.Minutes).To(Equal(5))
			Expect(glucoseRange.Variance).To(Equal(0.0))
		})

		It("range.UpdateTotal without minutes", func() {
			glucoseRange := Range{}

			By("adding 1 record of 5mmol")
			datum := NewGlucoseWithValue(selfmonitored.Type, now, 5)
			glucoseRange.UpdateTotal(datum)
			Expect(glucoseRange.Glucose).To(Equal(5.0))
			Expect(glucoseRange.Records).To(Equal(1))
			Expect(glucoseRange.Minutes).To(Equal(0))
			Expect(glucoseRange.Variance).To(Equal(0.0))

			By("adding 1 record of 10mmol")
			datum = NewGlucoseWithValue(selfmonitored.Type, now, 10)
			glucoseRange.UpdateTotal(datum)
			Expect(glucoseRange.Glucose).To(Equal(15.0))
			Expect(glucoseRange.Records).To(Equal(2))
			Expect(glucoseRange.Minutes).To(Equal(0))
			Expect(glucoseRange.Variance).To(Equal(0.0))
		})

		It("range.Update", func() {
			glucoseRange := Range{}
			datum := NewGlucoseWithValue(continuous.Type, now, 5)

			By("adding 5 minutes of 5mmol")
			glucoseRange.Update(datum)
			Expect(glucoseRange.Glucose).To(Equal(0.0))
			Expect(glucoseRange.Records).To(Equal(1))
			Expect(glucoseRange.Minutes).To(Equal(5))
			Expect(glucoseRange.Variance).To(Equal(0.0))
		})

		It("range.Update without minutes", func() {
			glucoseRange := Range{}

			By("adding 1 record of 5mmol")
			datum := NewGlucoseWithValue(selfmonitored.Type, now, 5)
			glucoseRange.Update(datum)
			Expect(glucoseRange.Glucose).To(Equal(0.0))
			Expect(glucoseRange.Records).To(Equal(1))
			Expect(glucoseRange.Minutes).To(Equal(0))
			Expect(glucoseRange.Variance).To(Equal(0.0))

			By("adding 1 record of 10mmol")
			datum = NewGlucoseWithValue(selfmonitored.Type, now, 10)
			glucoseRange.Update(datum)
			Expect(glucoseRange.Glucose).To(Equal(0.0))
			Expect(glucoseRange.Records).To(Equal(2))
			Expect(glucoseRange.Minutes).To(Equal(0))
			Expect(glucoseRange.Variance).To(Equal(0.0))
		})

		It("range.Add", func() {
			firstRange := Range{
				Glucose:  5,
				Minutes:  5,
				Records:  5,
				Percent:  5,
				Variance: 5,
			}

			secondRange := Range{
				Glucose:  10,
				Minutes:  10,
				Records:  10,
				Percent:  10,
				Variance: 10,
			}

			firstRange.Add(&secondRange)

			Expect(firstRange.Glucose).To(Equal(15.0))
			Expect(firstRange.Minutes).To(Equal(15))
			Expect(firstRange.Records).To(Equal(15))
			Expect(firstRange.Variance).To(Equal(15.0))

			// expect percent cleared, we don't handle percent on add
			Expect(firstRange.Percent).To(BeZero())
		})
	})

	Context("Ranges", func() {
		It("ranges.Add", func() {
			firstRanges := GlucoseRanges{
				Total: Range{
					Glucose:  10,
					Minutes:  11,
					Records:  12,
					Percent:  13,
					Variance: 14,
				},
				VeryLow: Range{
					Glucose:  20,
					Minutes:  21,
					Records:  22,
					Percent:  23,
					Variance: 24,
				},
				Low: Range{
					Glucose:  30,
					Minutes:  31,
					Records:  32,
					Percent:  33,
					Variance: 34,
				},
				Target: Range{
					Glucose:  40,
					Minutes:  41,
					Records:  42,
					Percent:  43,
					Variance: 44,
				},
				High: Range{
					Glucose:  50,
					Minutes:  51,
					Records:  52,
					Percent:  53,
					Variance: 54,
				},
				VeryHigh: Range{
					Glucose:  60,
					Minutes:  61,
					Records:  62,
					Percent:  63,
					Variance: 64,
				},
				ExtremeHigh: Range{
					Glucose:  70,
					Minutes:  71,
					Records:  72,
					Percent:  73,
					Variance: 74,
				},
				AnyLow: Range{
					Glucose:  80,
					Minutes:  81,
					Records:  82,
					Percent:  83,
					Variance: 84,
				},
				AnyHigh: Range{
					Glucose:  90,
					Minutes:  91,
					Records:  92,
					Percent:  93,
					Variance: 94,
				},
			}

			secondRanges := GlucoseRanges{
				Total: Range{
					Glucose:  15,
					Minutes:  16,
					Records:  17,
					Percent:  18,
					Variance: 19,
				},
				VeryLow: Range{
					Glucose:  25,
					Minutes:  26,
					Records:  27,
					Percent:  28,
					Variance: 29,
				},
				Low: Range{
					Glucose:  35,
					Minutes:  36,
					Records:  37,
					Percent:  38,
					Variance: 39,
				},
				Target: Range{
					Glucose:  45,
					Minutes:  46,
					Records:  47,
					Percent:  48,
					Variance: 49,
				},
				High: Range{
					Glucose:  55,
					Minutes:  56,
					Records:  57,
					Percent:  58,
					Variance: 59,
				},
				VeryHigh: Range{
					Glucose:  65,
					Minutes:  66,
					Records:  67,
					Percent:  68,
					Variance: 69,
				},
				ExtremeHigh: Range{
					Glucose:  75,
					Minutes:  76,
					Records:  77,
					Percent:  78,
					Variance: 79,
				},
				AnyLow: Range{
					Glucose:  85,
					Minutes:  86,
					Records:  87,
					Percent:  88,
					Variance: 89,
				},
				AnyHigh: Range{
					Glucose:  95,
					Minutes:  96,
					Records:  97,
					Percent:  98,
					Variance: 99,
				},
			}

			firstRanges.Add(&secondRanges)

			expectedRanges := GlucoseRanges{
				Total: Range{
					Glucose:  25,
					Minutes:  27,
					Records:  29,
					Percent:  0,
					Variance: 33.00526094276094,
				},
				VeryLow: Range{
					Glucose:  45,
					Minutes:  47,
					Records:  49,
					Percent:  0,
					Variance: 53.00097420310186,
				},
				Low: Range{
					Glucose:  65,
					Minutes:  67,
					Records:  69,
					Percent:  0,
					Variance: 73.0003343497566,
				},
				Target: Range{
					Glucose:  85,
					Minutes:  87,
					Records:  89,
					Percent:  0,
					Variance: 93.00015236284297,
				},
				High: Range{
					Glucose:  105,
					Minutes:  107,
					Records:  109,
					Percent:  0,
					Variance: 113.0000818084243,
				},
				VeryHigh: Range{
					Glucose:  125,
					Minutes:  127,
					Records:  129,
					Percent:  0,
					Variance: 133.00004889478234,
				},
				ExtremeHigh: Range{
					Glucose:  145,
					Minutes:  147,
					Records:  149,
					Percent:  0,
					Variance: 153.00003151742536,
				},
				AnyLow: Range{
					Glucose:  165,
					Minutes:  167,
					Records:  169,
					Percent:  0,
					Variance: 173.0000214901807,
				},
				AnyHigh: Range{
					Glucose:  185,
					Minutes:  187,
					Records:  189,
					Percent:  0,
					Variance: 193.00001530332412,
				},
			}

			Expect(firstRanges).To(BeComparableTo(expectedRanges))
		})

		It("ranges.Update", func() {
			glucoseRanges := GlucoseRanges{}

			By("adding a VeryLow value")
			glucoseRecord := NewGlucoseWithValue(continuous.Type, bucketTime, VeryLowBloodGlucose-0.1)
			Expect(glucoseRanges.Total.Records).To(Equal(0))
			Expect(glucoseRanges.VeryLow.Records).To(Equal(0))
			glucoseRanges.Update(glucoseRecord)
			Expect(glucoseRanges.VeryLow.Records).To(Equal(1))
			Expect(glucoseRanges.Total.Records).To(Equal(1))

			By("adding a Low value")
			glucoseRecord = NewGlucoseWithValue(continuous.Type, bucketTime, LowBloodGlucose-0.1)
			Expect(glucoseRanges.Low.Records).To(Equal(0))
			glucoseRanges.Update(glucoseRecord)
			Expect(glucoseRanges.Low.Records).To(Equal(1))
			Expect(glucoseRanges.Total.Records).To(Equal(2))

			By("adding a Target value")
			glucoseRecord = NewGlucoseWithValue(continuous.Type, bucketTime, InTargetBloodGlucose+0.1)
			Expect(glucoseRanges.Target.Records).To(Equal(0))
			glucoseRanges.Update(glucoseRecord)
			Expect(glucoseRanges.Target.Records).To(Equal(1))
			Expect(glucoseRanges.Total.Records).To(Equal(3))

			By("adding a High value")
			glucoseRecord = NewGlucoseWithValue(continuous.Type, bucketTime, HighBloodGlucose+0.1)
			Expect(glucoseRanges.High.Records).To(Equal(0))
			glucoseRanges.Update(glucoseRecord)
			Expect(glucoseRanges.High.Records).To(Equal(1))
			Expect(glucoseRanges.Total.Records).To(Equal(4))

			By("adding a VeryHigh value")
			glucoseRecord = NewGlucoseWithValue(continuous.Type, bucketTime, VeryHighBloodGlucose+0.1)
			Expect(glucoseRanges.VeryHigh.Records).To(Equal(0))
			glucoseRanges.Update(glucoseRecord)
			Expect(glucoseRanges.VeryHigh.Records).To(Equal(1))
			Expect(glucoseRanges.Total.Records).To(Equal(5))

			By("adding a High value")
			glucoseRecord = NewGlucoseWithValue(continuous.Type, bucketTime, ExtremeHighBloodGlucose+0.1)
			Expect(glucoseRanges.ExtremeHigh.Records).To(Equal(0))
			glucoseRanges.Update(glucoseRecord)
			Expect(glucoseRanges.ExtremeHigh.Records).To(Equal(1))
			Expect(glucoseRanges.VeryHigh.Records).To(Equal(2))
			Expect(glucoseRanges.Total.Records).To(Equal(6))
		})

		It("ranges.Finalize with minutes >70% of a day", func() {
			totalMinutes := 60.0 * 17.0
			glucoseRanges := GlucoseRanges{
				Total: Range{
					Minutes: int(totalMinutes),
					Records: 100,
				},
				VeryLow: Range{
					Minutes: 60 * 1,
					Records: 10,
				},
				Low: Range{
					Minutes: 60 * 2,
					Records: 20,
				},
				Target: Range{
					Minutes: 60 * 3,
					Records: 30,
				},
				High: Range{
					Minutes: 60 * 4,
					Records: 40,
				},
				VeryHigh: Range{
					Minutes: 60 * 5,
					Records: 50,
				},
				ExtremeHigh: Range{
					Minutes: 60 * 6,
					Records: 60,
				},
				AnyLow: Range{
					Minutes: 60 * 7,
					Records: 70,
				},
				AnyHigh: Range{
					Minutes: 60 * 8,
					Records: 80,
				},
			}

			glucoseRanges.Finalize(1)

			Expect(glucoseRanges.Total.Percent).To(Equal(17.0 / 24.0))
			Expect(glucoseRanges.VeryLow.Percent).To(Equal(60.0 * 1.0 / totalMinutes))
			Expect(glucoseRanges.Low.Percent).To(Equal(60.0 * 2.0 / totalMinutes))
			Expect(glucoseRanges.Target.Percent).To(Equal(60.0 * 3.0 / totalMinutes))
			Expect(glucoseRanges.High.Percent).To(Equal(60.0 * 4.0 / totalMinutes))
			Expect(glucoseRanges.VeryHigh.Percent).To(Equal(60.0 * 5.0 / totalMinutes))
			Expect(glucoseRanges.ExtremeHigh.Percent).To(Equal(60.0 * 6.0 / totalMinutes))
			Expect(glucoseRanges.AnyLow.Percent).To(Equal(60.0 * 7.0 / totalMinutes))
			Expect(glucoseRanges.AnyHigh.Percent).To(Equal(60.0 * 8.0 / totalMinutes))
		})

		It("ranges.Finalize with no minutes", func() {
			glucoseRanges := GlucoseRanges{
				Total: Range{
					Records: 100,
				},
				VeryLow: Range{
					Records: 10,
				},
				Low: Range{
					Records: 20,
				},
				Target: Range{
					Records: 30,
				},
				High: Range{
					Records: 40,
				},
				VeryHigh: Range{
					Records: 50,
				},
				ExtremeHigh: Range{
					Records: 60,
				},
				AnyLow: Range{
					Records: 70,
				},
				AnyHigh: Range{
					Records: 80,
				},
			}

			glucoseRanges.Finalize(1)

			Expect(glucoseRanges.Total.Percent).To(Equal(0.0))
			Expect(glucoseRanges.VeryLow.Percent).To(Equal(10.0 / 100.0))
			Expect(glucoseRanges.Low.Percent).To(Equal(20.0 / 100.0))
			Expect(glucoseRanges.Target.Percent).To(Equal(30.0 / 100.0))
			Expect(glucoseRanges.High.Percent).To(Equal(40.0 / 100.0))
			Expect(glucoseRanges.VeryHigh.Percent).To(Equal(50.0 / 100.0))
			Expect(glucoseRanges.ExtremeHigh.Percent).To(Equal(60.0 / 100.0))
			Expect(glucoseRanges.AnyLow.Percent).To(Equal(70.0 / 100.0))
			Expect(glucoseRanges.AnyHigh.Percent).To(Equal(80.0 / 100.0))
		})
	})

	Context("bucket.Update", func() {
		var userBucket *Bucket[*GlucoseBucket, GlucoseBucket]
		var cgmDatum data.Datum

		It("With a cgm value", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBucket = NewBucket[*GlucoseBucket](userId, bucketTime, SummaryTypeCGM)
			cgmDatum = NewGlucoseWithValue(continuous.Type, datumTime, InTargetBloodGlucose)

			err = userBucket.Update(cgmDatum)
			Expect(err).ToNot(HaveOccurred())

			Expect(userBucket.FirstData).To(Equal(datumTime))
			Expect(userBucket.LastData).To(Equal(datumTime))
			Expect(userBucket.Time).To(Equal(bucketTime))
			Expect(userBucket.Data.Target.Records).To(Equal(1))
			Expect(userBucket.Data.Target.Minutes).To(Equal(5))
			Expect(userBucket.Data.Min).To(Equal(InTargetBloodGlucose))
			Expect(userBucket.Data.Max).To(Equal(InTargetBloodGlucose))
			Expect(userBucket.IsModified()).To(BeTrue())
		})

		It("With a bgm value", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBucket = NewBucket[*GlucoseBucket](userId, bucketTime, SummaryTypeBGM)
			bgmDatum := NewGlucoseWithValue(selfmonitored.Type, datumTime, InTargetBloodGlucose)

			err = userBucket.Update(bgmDatum)
			Expect(err).ToNot(HaveOccurred())

			Expect(userBucket.FirstData).To(Equal(datumTime))
			Expect(userBucket.LastData).To(Equal(datumTime))
			Expect(userBucket.Time).To(Equal(bucketTime))
			Expect(userBucket.Data.Target.Records).To(Equal(1))
			Expect(userBucket.Data.Target.Minutes).To(Equal(0))
			Expect(userBucket.Data.Min).To(Equal(InTargetBloodGlucose))
			Expect(userBucket.Data.Max).To(Equal(InTargetBloodGlucose))
			Expect(userBucket.IsModified()).To(BeTrue())
		})

		It("With two cgm values within 5 minutes", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBucket = NewBucket[*GlucoseBucket](userId, bucketTime, SummaryTypeCGM)
			cgmDatum = NewGlucoseWithValue(continuous.Type, datumTime, InTargetBloodGlucose)

			err = userBucket.Update(cgmDatum)
			Expect(err).ToNot(HaveOccurred())

			Expect(userBucket.FirstData).To(Equal(datumTime))
			Expect(userBucket.LastData).To(Equal(datumTime))
			Expect(userBucket.Data.Target.Records).To(Equal(1))
			Expect(userBucket.Data.Target.Minutes).To(Equal(5))
			Expect(userBucket.Data.Min).To(Equal(InTargetBloodGlucose))
			Expect(userBucket.Data.Max).To(Equal(InTargetBloodGlucose))

			newDatumTime := bucketTime.Add(9 * time.Minute)
			cgmDatum = NewGlucoseWithValue(continuous.Type, newDatumTime, InTargetBloodGlucose)
			err = userBucket.Update(cgmDatum)
			Expect(err).ToNot(HaveOccurred())

			Expect(userBucket.FirstData).To(Equal(datumTime))
			Expect(userBucket.LastData).To(Equal(datumTime))
			Expect(userBucket.Data.Target.Records).To(Equal(1))
			Expect(userBucket.Data.Target.Minutes).To(Equal(5))
			Expect(userBucket.Data.Min).To(Equal(InTargetBloodGlucose))
			Expect(userBucket.Data.Max).To(Equal(InTargetBloodGlucose))
		})

		It("With two bgm values within 5 minutes", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBucket = NewBucket[*GlucoseBucket](userId, bucketTime, SummaryTypeBGM)
			bgmDatum := NewGlucoseWithValue(selfmonitored.Type, datumTime, InTargetBloodGlucose)

			err = userBucket.Update(bgmDatum)
			Expect(err).ToNot(HaveOccurred())

			Expect(userBucket.LastData).To(Equal(datumTime))
			Expect(userBucket.FirstData).To(Equal(datumTime))
			Expect(userBucket.Data.Target.Records).To(Equal(1))
			Expect(userBucket.Data.Target.Minutes).To(Equal(0))
			Expect(userBucket.Data.Min).To(Equal(InTargetBloodGlucose))
			Expect(userBucket.Data.Max).To(Equal(InTargetBloodGlucose))

			newDatumTime := bucketTime.Add(9 * time.Minute)
			bgmDatum = NewGlucoseWithValue(selfmonitored.Type, newDatumTime, InTargetBloodGlucose)
			err = userBucket.Update(bgmDatum)
			Expect(err).ToNot(HaveOccurred())

			Expect(userBucket.FirstData).To(Equal(datumTime))
			Expect(userBucket.LastData).To(Equal(newDatumTime))
			Expect(userBucket.Data.Target.Records).To(Equal(2))
			Expect(userBucket.Data.Target.Minutes).To(Equal(0))
			Expect(userBucket.Data.Min).To(Equal(InTargetBloodGlucose))
			Expect(userBucket.Data.Max).To(Equal(InTargetBloodGlucose))
		})

		It("With two values in a range", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBucket = NewBucket[*GlucoseBucket](userId, bucketTime, SummaryTypeCGM)

			By("Inserting the first data")

			cgmDatum = NewGlucoseWithValue(continuous.Type, datumTime, InTargetBloodGlucose)
			err = userBucket.Update(cgmDatum)
			Expect(err).ToNot(HaveOccurred())

			Expect(userBucket.FirstData).To(Equal(datumTime))
			Expect(userBucket.LastData).To(Equal(datumTime))
			Expect(userBucket.Time).To(Equal(bucketTime))
			Expect(userBucket.Data.Target.Records).To(Equal(1))
			Expect(userBucket.Data.Target.Minutes).To(Equal(5))
			Expect(userBucket.Data.Min).To(Equal(InTargetBloodGlucose))
			Expect(userBucket.Data.Max).To(Equal(InTargetBloodGlucose))
			Expect(userBucket.IsModified()).To(BeTrue())

			secondDatumTime := datumTime.Add(5 * time.Minute)
			cgmDatum = NewGlucoseWithValue(continuous.Type, secondDatumTime, InTargetBloodGlucose)

			By("Inserting the second data")

			err = userBucket.Update(cgmDatum)
			Expect(err).ToNot(HaveOccurred())

			Expect(userBucket.FirstData).To(Equal(datumTime))
			Expect(userBucket.LastData).To(Equal(secondDatumTime))
			Expect(userBucket.Time).To(Equal(bucketTime))
			Expect(userBucket.Data.Target.Records).To(Equal(2))
			Expect(userBucket.Data.Target.Minutes).To(Equal(10))
			Expect(userBucket.Data.Min).To(Equal(InTargetBloodGlucose))
			Expect(userBucket.Data.Max).To(Equal(InTargetBloodGlucose))
			Expect(userBucket.IsModified()).To(BeTrue())
		})

		It("With values in all ranges", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBucket = NewBucket[*GlucoseBucket](userId, bucketTime, SummaryTypeCGM)

			ranges := map[float64]*Range{
				VeryLowBloodGlucose - 0.1:     &userBucket.Data.VeryLow,
				LowBloodGlucose - 0.1:         &userBucket.Data.Low,
				InTargetBloodGlucose + 0.1:    &userBucket.Data.Target,
				HighBloodGlucose + 0.1:        &userBucket.Data.High,
				ExtremeHighBloodGlucose + 0.1: &userBucket.Data.ExtremeHigh,
			}

			expectedMin := 9999.9
			expectedMax := -1.0

			expectedGlucose := 0.0
			expectedMinutes := 0
			expectedRecords := 0

			expectedAnyLowGlucose := 0.0
			expectedAnyLowMinutes := 0
			expectedAnyLowRecords := 0

			expectedAnyHighGlucose := 0.0
			expectedAnyHighMinutes := 0
			expectedAnyHighRecords := 0

			expectedVeryHighGlucose := 0.0
			expectedVeryHighMinutes := 0
			expectedVeryHighRecords := 0

			offset := 0
			for k, v := range ranges {
				By(fmt.Sprintf("Add a value of %f", k))
				Expect(v.Records).To(BeZero())
				Expect(v.Glucose).To(BeZero())
				Expect(v.Minutes).To(BeZero())

				cgmDatum = NewGlucoseWithValue(continuous.Type, datumTime.Add(5*time.Minute*time.Duration(offset)), k)
				err = userBucket.Update(cgmDatum)
				Expect(err).ToNot(HaveOccurred())

				Expect(v.Records).To(Equal(1))
				Expect(v.Minutes).To(Equal(5))

				expectedGlucose += k * 5
				expectedMinutes += 5
				expectedRecords++
				Expect(userBucket.Data.Total.Records).To(Equal(expectedRecords))
				Expect(userBucket.Data.Total.Glucose).To(Equal(expectedGlucose))
				Expect(userBucket.Data.Total.Minutes).To(Equal(expectedMinutes))

				if k < LowBloodGlucose {
					expectedAnyLowGlucose += k * 5
					expectedAnyLowMinutes += 5
					expectedAnyLowRecords++
				}
				Expect(userBucket.Data.AnyLow.Records).To(Equal(expectedAnyLowRecords))
				Expect(userBucket.Data.AnyLow.Minutes).To(Equal(expectedAnyLowMinutes))

				if k > HighBloodGlucose {
					expectedAnyHighGlucose += k * 5
					expectedAnyHighMinutes += 5
					expectedAnyHighRecords++
				}
				Expect(userBucket.Data.AnyHigh.Records).To(Equal(expectedAnyHighRecords))
				Expect(userBucket.Data.AnyHigh.Minutes).To(Equal(expectedAnyHighMinutes))

				if k > VeryHighBloodGlucose {
					expectedVeryHighGlucose += k * 5
					expectedVeryHighMinutes += 5
					expectedVeryHighRecords++
				}
				Expect(userBucket.Data.VeryHigh.Records).To(Equal(expectedVeryHighRecords))
				Expect(userBucket.Data.VeryHigh.Minutes).To(Equal(expectedVeryHighMinutes))

				if k > float64(expectedMax) {
					expectedMax = k
				}
				if k < float64(expectedMin) {
					expectedMin = k
				}
				Expect(userBucket.Data.Min).To(Equal(expectedMin))
				Expect(userBucket.Data.Max).To(Equal(expectedMax))

				if offset > 0 {
					Expect(userBucket.Data.Total.Variance).ToNot(BeZero())
				}

				offset++
			}
		})
	})

	Context("period", func() {
		var period GlucosePeriod

		It("Add single bucket to an empty period", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			period = GlucosePeriod{}

			bucketOne := NewBucket[*GlucoseBucket](userId, bucketTime, SummaryTypeCGM)
			err = bucketOne.Update(NewGlucoseWithValue(continuous.Type, datumTime, InTargetBloodGlucose))
			Expect(err).ToNot(HaveOccurred())

			err = period.Update(bucketOne)
			Expect(err).ToNot(HaveOccurred())
			Expect(period.Target.Records).To(Equal(1))
		})

		It("Add duplicate buckets to a period", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			period = GlucosePeriod{}

			bucketOne := NewBucket[*GlucoseBucket](userId, bucketTime, SummaryTypeCGM)
			err = bucketOne.Update(NewGlucoseWithValue(continuous.Type, datumTime, InTargetBloodGlucose))
			Expect(err).ToNot(HaveOccurred())

			err = period.Update(bucketOne)
			Expect(err).ToNot(HaveOccurred())
			Expect(period.Target.Records).To(Equal(1))

			err = period.Update(bucketOne)
			Expect(err).To(HaveOccurred())
		})

		It("Add three buckets to an empty period on 2 different days, 3 different hours", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			period = GlucosePeriod{}

			bucketOne := NewBucket[*GlucoseBucket](userId, bucketTime, SummaryTypeCGM)
			err = bucketOne.Update(NewGlucoseWithValue(continuous.Type, datumTime, InTargetBloodGlucose))
			Expect(err).ToNot(HaveOccurred())

			bucketTwo := NewBucket[*GlucoseBucket](userId, bucketTime.Add(time.Hour), SummaryTypeCGM)
			err = bucketTwo.Update(NewGlucoseWithValue(continuous.Type, datumTime.Add(time.Hour), InTargetBloodGlucose))
			Expect(err).ToNot(HaveOccurred())

			bucketThree := NewBucket[*GlucoseBucket](userId, bucketTime.Add(24*time.Hour), SummaryTypeCGM)
			err = bucketThree.Update(NewGlucoseWithValue(continuous.Type, datumTime.Add(24*time.Hour), InTargetBloodGlucose))
			Expect(err).ToNot(HaveOccurred())

			err = period.Update(bucketOne)
			Expect(err).ToNot(HaveOccurred())
			Expect(period.Target.Records).To(Equal(1))
			Expect(period.HoursWithData).To(Equal(1))
			Expect(period.DaysWithData).To(Equal(1))
			Expect(period.Min).To(Equal(InTargetBloodGlucose))
			Expect(period.Max).To(Equal(InTargetBloodGlucose))

			err = period.Update(bucketTwo)
			Expect(err).ToNot(HaveOccurred())
			Expect(period.Target.Records).To(Equal(2))
			Expect(period.HoursWithData).To(Equal(2))
			Expect(period.DaysWithData).To(Equal(1))
			Expect(period.Min).To(Equal(InTargetBloodGlucose))
			Expect(period.Max).To(Equal(InTargetBloodGlucose))

			err = period.Update(bucketThree)
			Expect(err).ToNot(HaveOccurred())
			Expect(period.Target.Records).To(Equal(3))
			Expect(period.HoursWithData).To(Equal(3))
			Expect(period.DaysWithData).To(Equal(2))
			Expect(period.Min).To(Equal(InTargetBloodGlucose))
			Expect(period.Max).To(Equal(InTargetBloodGlucose))
		})

		It("Finalize a 1d period", func() {
			period = GlucosePeriod{}
			buckets := CreateGlucoseBuckets(bucketTime, 24, 12, true)

			for i := range buckets {
				err = period.Update(buckets[i])
				Expect(err).ToNot(HaveOccurred())
			}

			period.Finalize(1)

			// data is generated at 100% per range
			Expect(period.VeryHigh.Percent).To(Equal(1.0))
			Expect(period.AnyLow.Percent).To(Equal(1.0))
			Expect(period.AnyHigh.Percent).To(Equal(1.0))
			Expect(period.Target.Percent).To(Equal(1.0))
			Expect(period.Low.Percent).To(Equal(1.0))
			Expect(period.High.Percent).To(Equal(1.0))
			Expect(period.VeryLow.Percent).To(Equal(1.0))
			Expect(period.ExtremeHigh.Percent).To(Equal(1.0))

			Expect(period.AverageDailyRecords).To(Equal(12.0 * 24.0))
			Expect(period.AverageGlucose).To(Equal(InTargetBloodGlucose))
			Expect(period.GlucoseManagementIndicator).To(Equal(CalculateGMI(InTargetBloodGlucose)))

			// we only validate these are set here, as this requires more specific validation
			Expect(period.StandardDeviation).ToNot(Equal(0.0))
			Expect(period.CoefficientOfVariation).ToNot(Equal(0.0))
		})

		It("Finalize a 7d period", func() {
			period = GlucosePeriod{}
			buckets := CreateGlucoseBuckets(bucketTime, 24*5, 12, true)

			for i := range buckets {
				err = period.Update(buckets[i])
				Expect(err).ToNot(HaveOccurred())
			}

			period.Finalize(7)

			// data is generated at 100% per range
			Expect(period.VeryHigh.Percent).To(Equal(1.0))
			Expect(period.AnyLow.Percent).To(Equal(1.0))
			Expect(period.AnyHigh.Percent).To(Equal(1.0))
			Expect(period.Target.Percent).To(Equal(1.0))
			Expect(period.Low.Percent).To(Equal(1.0))
			Expect(period.High.Percent).To(Equal(1.0))
			Expect(period.VeryLow.Percent).To(Equal(1.0))
			Expect(period.ExtremeHigh.Percent).To(Equal(1.0))

			Expect(period.AverageDailyRecords).To(Equal((12.0 * 24.0) * 5 / 7))
			Expect(period.AverageGlucose).To(Equal(InTargetBloodGlucose))
			Expect(period.GlucoseManagementIndicator).To(Equal(CalculateGMI(InTargetBloodGlucose)))

			// we only validate these are set here, as this requires more specific validation
			Expect(period.StandardDeviation).ToNot(Equal(0.0))
			Expect(period.CoefficientOfVariation).ToNot(Equal(0.0))
		})

		It("Finalize a 1d period with insufficient data", func() {
			period = GlucosePeriod{}
			buckets := CreateGlucoseBuckets(bucketTime, 16, 12, true)

			for i := range buckets {
				err = period.Update(buckets[i])
				Expect(err).ToNot(HaveOccurred())
			}

			period.Finalize(1)

			// data is generated at 100% per range
			Expect(period.VeryHigh.Percent).To(Equal(1.0))
			Expect(period.AnyLow.Percent).To(Equal(1.0))
			Expect(period.AnyHigh.Percent).To(Equal(1.0))
			Expect(period.Target.Percent).To(Equal(1.0))
			Expect(period.Low.Percent).To(Equal(1.0))
			Expect(period.High.Percent).To(Equal(1.0))
			Expect(period.VeryLow.Percent).To(Equal(1.0))
			Expect(period.ExtremeHigh.Percent).To(Equal(1.0))

			Expect(period.AverageDailyRecords).To(Equal(12.0 * 16.0))
			Expect(period.AverageGlucose).To(Equal(InTargetBloodGlucose))
			Expect(period.GlucoseManagementIndicator).To(Equal(5.5))

			// we only validate these are set here, as this requires more specific validation
			Expect(period.StandardDeviation).ToNot(Equal(0.0))
			Expect(period.CoefficientOfVariation).ToNot(Equal(0.0))
		})

		It("Finalize a 7d period with insufficient data", func() {
			period = GlucosePeriod{}
			buckets := CreateGlucoseBuckets(bucketTime, 23, 12, true)

			for i := range buckets {
				err = period.Update(buckets[i])
				Expect(err).ToNot(HaveOccurred())
			}

			period.Finalize(7)

			// data is generated at 100% per range
			Expect(period.VeryHigh.Percent).To(Equal(1.0))
			Expect(period.AnyLow.Percent).To(Equal(1.0))
			Expect(period.AnyHigh.Percent).To(Equal(1.0))
			Expect(period.Target.Percent).To(Equal(1.0))
			Expect(period.Low.Percent).To(Equal(1.0))
			Expect(period.High.Percent).To(Equal(1.0))
			Expect(period.VeryLow.Percent).To(Equal(1.0))
			Expect(period.ExtremeHigh.Percent).To(Equal(1.0))

			Expect(period.AverageDailyRecords).To(Equal(12.0 * 23.0 / 7))
			Expect(period.AverageGlucose).To(Equal(InTargetBloodGlucose))
			Expect(period.GlucoseManagementIndicator).To(Equal(5.5))

			// we only validate these are set here, as this requires more specific validation
			Expect(period.StandardDeviation).ToNot(Equal(0.0))
			Expect(period.CoefficientOfVariation).ToNot(Equal(0.0))
		})

		It("Update a finalized period", func() {
			period = GlucosePeriod{}
			period.Finalize(14)

			bucket := NewBucket[*GlucoseBucket](userId, bucketTime, SummaryTypeCGM)
			err = period.Update(bucket)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("GlucosePeriods", func() {
		var logger log.Logger
		var ctx context.Context

		BeforeEach(func() {
			logger = logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), logger)
		})

		It("Init", func() {
			s := GlucosePeriods{}
			s.Init()

			Expect(s).ToNot(BeNil())
		})

		Context("Update", func() {

			It("Update 1d", func() {
				s := GlucosePeriods{}
				s.Init()

				buckets := CreateGlucoseBuckets(bucketTime, 24, 1, true)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.Update(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s["1d"].Total.Records).To(Equal(24))
				Expect(s["7d"].Total.Records).To(Equal(24))
				Expect(s["14d"].Total.Records).To(Equal(24))
				Expect(s["30d"].Total.Records).To(Equal(24))
			})

			It("CalculateSummary 2d", func() {
				s := GlucosePeriods{}
				s.Init()

				buckets := CreateGlucoseBuckets(bucketTime, 48, 1, true)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.Update(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s["1d"].Total.Records).To(Equal(24))
				Expect(s["7d"].Total.Records).To(Equal(24 * 2))
				Expect(s["14d"].Total.Records).To(Equal(24 * 2))
				Expect(s["30d"].Total.Records).To(Equal(24 * 2))
			})

			It("CalculateSummary 7d", func() {
				s := GlucosePeriods{}
				s.Init()

				buckets := CreateGlucoseBuckets(bucketTime, 24*7, 1, true)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.Update(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s["1d"].Total.Records).To(Equal(24))
				Expect(s["7d"].Total.Records).To(Equal(24 * 7))
				Expect(s["14d"].Total.Records).To(Equal(24 * 7))
				Expect(s["30d"].Total.Records).To(Equal(24 * 7))
			})

			It("CalculateSummary 14d", func() {
				s := GlucosePeriods{}
				s.Init()

				buckets := CreateGlucoseBuckets(bucketTime, 24*14, 1, true)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.Update(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s["1d"].Total.Records).To(Equal(24))
				Expect(s["7d"].Total.Records).To(Equal(24 * 7))
				Expect(s["14d"].Total.Records).To(Equal(24 * 14))
				Expect(s["30d"].Total.Records).To(Equal(24 * 14))
			})

			It("CalculateSummary 28d", func() {
				s := GlucosePeriods{}
				s.Init()

				buckets := CreateGlucoseBuckets(bucketTime, 24*28, 1, true)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.Update(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s["1d"].Total.Records).To(Equal(24))
				Expect(s["7d"].Total.Records).To(Equal(24 * 7))
				Expect(s["14d"].Total.Records).To(Equal(24 * 14))
				Expect(s["30d"].Total.Records).To(Equal(24 * 28))
			})

			It("CalculateSummary 30d", func() {
				s := GlucosePeriods{}
				s.Init()

				buckets := CreateGlucoseBuckets(bucketTime, 24*30, 1, true)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.Update(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s["1d"].Total.Records).To(Equal(24))
				Expect(s["7d"].Total.Records).To(Equal(24 * 7))
				Expect(s["14d"].Total.Records).To(Equal(24 * 14))
				Expect(s["30d"].Total.Records).To(Equal(24 * 30))
			})

			It("CalculateSummary 60d", func() {
				s := GlucosePeriods{}
				s.Init()

				buckets := CreateGlucoseBuckets(bucketTime, 24*60, 1, true)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.Update(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s["1d"].Total.Records).To(Equal(24))
				Expect(s["7d"].Total.Records).To(Equal(24 * 7))
				Expect(s["14d"].Total.Records).To(Equal(24 * 14))
				Expect(s["30d"].Total.Records).To(Equal(24 * 30))
			})

			It("CalculateSummary 61d", func() {
				s := GlucosePeriods{}
				s.Init()

				buckets := CreateGlucoseBuckets(bucketTime, 24*61, 1, true)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.Update(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s["1d"].Total.Records).To(Equal(24))
				Expect(s["7d"].Total.Records).To(Equal(24 * 7))
				Expect(s["14d"].Total.Records).To(Equal(24 * 14))
				Expect(s["30d"].Total.Records).To(Equal(24 * 30))
			})
		})

		Context("CalculateDelta", func() {

			It("CalculateDelta populates all values", func() {
				// This validates a large block of easy to typo function calls in CalculateDelta, apologies to whoever has
				// to update this.
				s := GlucosePeriods{"1d": &GlucosePeriod{
					GlucoseRanges: GlucoseRanges{
						Total: Range{
							Glucose:  0,
							Minutes:  0,
							Records:  0,
							Percent:  0,
							Variance: 0,
						},
						VeryLow: Range{
							Glucose:  1,
							Minutes:  1,
							Records:  1,
							Percent:  1,
							Variance: 1,
						},
						Low: Range{
							Glucose:  2,
							Minutes:  2,
							Records:  2,
							Percent:  2,
							Variance: 2,
						},
						Target: Range{
							Glucose:  3,
							Minutes:  3,
							Records:  3,
							Percent:  3,
							Variance: 3,
						},
						High: Range{
							Glucose:  4,
							Minutes:  4,
							Records:  4,
							Percent:  4,
							Variance: 4,
						},
						VeryHigh: Range{
							Glucose:  5,
							Minutes:  5,
							Records:  5,
							Percent:  5,
							Variance: 5,
						},
						ExtremeHigh: Range{
							Glucose:  6,
							Minutes:  6,
							Records:  6,
							Percent:  6,
							Variance: 6,
						},
						AnyLow: Range{
							Glucose:  7,
							Minutes:  7,
							Records:  7,
							Percent:  7,
							Variance: 7,
						},
						AnyHigh: Range{
							Glucose:  8,
							Minutes:  8,
							Records:  8,
							Percent:  8,
							Variance: 8,
						},
					},
					MinMax:                     MinMax{Min: 3, Max: 5},
					HoursWithData:              0,
					DaysWithData:               1,
					AverageGlucose:             2,
					GlucoseManagementIndicator: 3,
					CoefficientOfVariation:     4,
					StandardDeviation:          5,
					AverageDailyRecords:        6,
				},
				}
				offset := GlucosePeriods{"1d": &GlucosePeriod{
					GlucoseRanges: GlucoseRanges{
						Total: Range{
							Glucose:  99,
							Minutes:  98,
							Records:  97,
							Percent:  96,
							Variance: 95,
						},
						VeryLow: Range{
							Glucose:  89,
							Minutes:  88,
							Records:  87,
							Percent:  86,
							Variance: 85,
						},
						Low: Range{
							Glucose:  79,
							Minutes:  78,
							Records:  77,
							Percent:  76,
							Variance: 75,
						},
						Target: Range{
							Glucose:  69,
							Minutes:  68,
							Records:  67,
							Percent:  66,
							Variance: 65,
						},
						High: Range{
							Glucose:  59,
							Minutes:  58,
							Records:  57,
							Percent:  56,
							Variance: 55,
						},
						VeryHigh: Range{
							Glucose:  49,
							Minutes:  48,
							Records:  47,
							Percent:  46,
							Variance: 45,
						},
						ExtremeHigh: Range{
							Glucose:  39,
							Minutes:  38,
							Records:  37,
							Percent:  36,
							Variance: 35,
						},
						AnyLow: Range{
							Glucose:  29,
							Minutes:  28,
							Records:  27,
							Percent:  26,
							Variance: 25,
						},
						AnyHigh: Range{
							Glucose:  19,
							Minutes:  18,
							Records:  17,
							Percent:  16,
							Variance: 15,
						},
					},
					MinMax:                     MinMax{Min: 5, Max: 6},
					HoursWithData:              99,
					DaysWithData:               98,
					AverageGlucose:             97,
					GlucoseManagementIndicator: 96,
					CoefficientOfVariation:     95,
					StandardDeviation:          94,
					AverageDailyRecords:        93,
				},
				}

				s.CalculateDelta(offset)

				expectedDelta := GlucosePeriod{
					GlucoseRanges: GlucoseRanges{
						Total: Range{
							Minutes: -98,
							Records: -97,
							Percent: -96,
						},
						VeryLow: Range{
							Minutes: -87,
							Records: -86,
							Percent: -85,
						},
						Low: Range{
							Minutes: -76,
							Records: -75,
							Percent: -74,
						},
						Target: Range{
							Minutes: -65,
							Records: -64,
							Percent: -63,
						},
						High: Range{
							Minutes: -54,
							Records: -53,
							Percent: -52,
						},
						VeryHigh: Range{
							Minutes: -43,
							Records: -42,
							Percent: -41,
						},
						ExtremeHigh: Range{
							Minutes: -32,
							Records: -31,
							Percent: -30,
						},
						AnyLow: Range{
							Minutes: -21,
							Records: -20,
							Percent: -19,
						},
						AnyHigh: Range{
							Minutes: -10,
							Records: -9,
							Percent: -8,
						},
					},
					MinMax:                     MinMax{Min: -2, Max: -1},
					HoursWithData:              -99,
					DaysWithData:               -97,
					AverageGlucose:             -95,
					GlucoseManagementIndicator: -93,
					CoefficientOfVariation:     -91,
					StandardDeviation:          -89,
					AverageDailyRecords:        -87,
				}

				opts := cmpopts.IgnoreUnexported(GlucosePeriod{})
				Expect(*(s["1d"].Delta)).To(BeComparableTo(expectedDelta, opts))

			})

			It("CalculateDelta 1d", func() {
				s := GlucosePeriods{}
				s.Init()

				bucketsOne := CreateGlucoseBuckets(bucketTime, 24, 1, true)
				bucketsTwo := CreateGlucoseBuckets(bucketTime.AddDate(0, 0, -1), 24, 2, true)
				buckets := append(bucketsOne, bucketsTwo...)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.Update(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s["1d"].Delta.Total.Records).To(Equal(-24))
			})

			It("CalculateDelta 7d", func() {
				s := GlucosePeriods{}
				s.Init()

				bucketsOne := CreateGlucoseBuckets(bucketTime, 24*7, 1, true)
				bucketsTwo := CreateGlucoseBuckets(bucketTime.AddDate(0, 0, -7), 24*7, 2, true)
				buckets := append(bucketsOne, bucketsTwo...)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.Update(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s["7d"].Delta.Total.Records).To(Equal(-24 * 7))
			})

			It("CalculateDelta 14d", func() {
				s := GlucosePeriods{}
				s.Init()

				bucketsOne := CreateGlucoseBuckets(bucketTime, 24*14, 1, true)
				bucketsTwo := CreateGlucoseBuckets(bucketTime.AddDate(0, 0, -14), 24*14, 2, true)
				buckets := append(bucketsOne, bucketsTwo...)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.Update(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s["14d"].Delta.Total.Records).To(Equal(-24 * 14))
			})

			It("CalculateDelta 30d", func() {
				s := GlucosePeriods{}
				s.Init()

				bucketsOne := CreateGlucoseBuckets(bucketTime, 24*30, 1, true)
				bucketsTwo := CreateGlucoseBuckets(bucketTime.AddDate(0, 0, -30), 24*30, 2, true)
				buckets := append(bucketsOne, bucketsTwo...)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.Update(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s["30d"].Delta.Total.Records).To(Equal(-24 * 30))
			})
		})
	})
})
