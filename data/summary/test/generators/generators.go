package generators

import (
	"github.com/tidepool-org/platform/data/summary/types"
	"math"
	"math/rand/v2"
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/pointer"
)

const (
	VeryLowBloodGlucose     = 3.0
	LowBloodGlucose         = 3.9
	HighBloodGlucose        = 10.0
	VeryHighBloodGlucose    = 13.9
	ExtremeHighBloodGlucose = 19.4
	InTargetBloodGlucose    = 5.0
)

func ConvertToIntArray[T data.Datum](arr []T) []interface{} {
	s := make([]interface{}, len(arr))
	for i, v := range arr {
		s[i] = v
	}

	return s
}

func ExpectedAverage(windowSize int, hoursAdded int, newAvg float64, oldAvg float64) float64 {
	oldHoursRemaining := windowSize - hoursAdded
	oldAvgTotal := oldAvg * math.Max(float64(oldHoursRemaining), 0)
	newAvgTotal := newAvg * math.Min(float64(hoursAdded), float64(windowSize))

	return (oldAvgTotal + newAvgTotal) / float64(windowSize)
}

var Units = "mmol/L"

type DataRanges struct {
	Min         float64
	Max         float64
	Padding     float64
	VeryLow     float64
	Low         float64
	High        float64
	VeryHigh    float64
	ExtremeHigh float64
}

func NewDataRanges() DataRanges {
	return DataRanges{
		Min:         1,
		Max:         25,
		Padding:     0.01,
		VeryLow:     VeryLowBloodGlucose,
		Low:         LowBloodGlucose,
		High:        HighBloodGlucose,
		VeryHigh:    VeryHighBloodGlucose,
		ExtremeHigh: ExtremeHighBloodGlucose,
	}
}

func NewDataRangesSingle(value float64) DataRanges {
	return DataRanges{
		Min:         value,
		Max:         value,
		Padding:     0,
		VeryLow:     value,
		Low:         value,
		High:        value,
		VeryHigh:    value,
		ExtremeHigh: value,
	}
}

func Mean(x []float64) float64 {
	sum := 0.0
	for i := 0; i < len(x); i++ {
		sum += x[i]
	}
	return sum / float64(len(x))
}

func calcVariance(x []float64, mean float64) float64 {
	var (
		ss           float64
		compensation float64
	)
	for _, v := range x {
		d := v - mean
		ss += d * d
		compensation += d
	}
	return ss - compensation*compensation/float64(len(x))
}

func PopStdDev(x []float64) float64 {
	variance := calcVariance(x, Mean(x)) / float64(len(x))
	return math.Sqrt(variance)
}

func NewGlucose(typ *string, units *string, datumTime *time.Time, deviceID *string, uploadId *string) *glucose.Glucose {
	timestamp := time.Now().UTC().Truncate(time.Millisecond)
	datum := glucose.New(*typ)
	datum.Units = units

	datum.Active = true
	datum.ArchivedDataSetID = nil
	datum.ArchivedTime = nil
	datum.CreatedTime = &timestamp
	datum.CreatedUserID = nil
	datum.DeletedTime = nil
	datum.DeletedUserID = nil
	datum.DeviceID = deviceID
	datum.ModifiedTime = &timestamp
	datum.ModifiedUserID = nil
	datum.Time = datumTime
	datum.UploadID = uploadId

	return &datum
}

func NewGlucoseWithValue(typ string, datumTime time.Time, value float64) (g *glucose.Glucose) {
	g = NewGlucose(&typ, &Units, &datumTime, pointer.FromAny("SummaryTestDevice"), pointer.FromAny(test.RandomSetID()))
	g.Value = &value
	return
}

func NewDataSetCGMDataAvg(startTime time.Time, hours float64, reqAvg float64) []data.Datum {
	requiredRecords := int(hours * 12)
	typ := "cbg"
	dataSetData := make([]data.Datum, requiredRecords)
	deviceId := "SummaryTestDevice"
	uploadId := test.RandomSetID()

	// generate X hours of data
	for count := 0; count < requiredRecords; count += 2 {
		randValue := 1 + rand.Float64()*(reqAvg-1)
		glucoseValues := [2]float64{reqAvg + randValue, reqAvg - randValue}

		// this adds 2 entries, one for each side of the average so that the calculated average is the requested value
		for i, glucoseValue := range glucoseValues {
			datumTime := startTime.Add(time.Duration(-(count + i + 1)) * time.Minute * 5)

			datum := NewGlucose(&typ, &Units, &datumTime, &deviceId, &uploadId)
			datum.Value = pointer.FromAny(glucoseValue)

			dataSetData[requiredRecords-count-i-1] = datum
		}
	}

	return dataSetData
}

// creates a dataset with random values evenly divided between ranges
func NewDataSetCGMDataRanges(startTime time.Time, hours float64, ranges DataRanges) []data.Datum {
	perHour := 12.0
	requiredRecords := int(hours * perHour)
	typ := "cbg"
	dataSetData := make([]data.Datum, requiredRecords)
	uploadId := test.RandomSetID()
	deviceId := "SummaryTestDevice"

	glucoseBrackets := [6][2]float64{
		{ranges.Min, ranges.VeryLow - ranges.Padding},
		{ranges.VeryLow, ranges.Low - ranges.Padding},
		{ranges.Low, ranges.High - ranges.Padding},
		{ranges.High, ranges.VeryHigh - ranges.Padding},
		{ranges.VeryHigh, ranges.ExtremeHigh - ranges.Padding},
		{ranges.ExtremeHigh, ranges.Max},
	}

	// generate requiredRecords of data
	for count := 0; count < requiredRecords; count += 6 {
		for i, bracket := range glucoseBrackets {
			datumTime := startTime.Add(time.Duration(-(count + i + 1)) * time.Minute * 5)

			datum := NewGlucose(&typ, &Units, &datumTime, &deviceId, &uploadId)
			datum.Value = pointer.FromAny(bracket[0] + (bracket[1]-bracket[0])*rand.Float64())

			dataSetData[requiredRecords-count-i-1] = datum
		}
	}

	return dataSetData
}

func NewDataSetCGMVariance(startTime time.Time, hours int, perHour int, StandardDeviation float64) ([]data.Datum, float64) {
	requiredRecords := hours * perHour
	typ := "cbg"
	dataSetData := make([]data.Datum, requiredRecords)
	uploadId := test.RandomSetID()
	deviceId := "SummaryTestDevice"

	var values []float64

	// generate requiredRecords of data
	for count := 0; count < requiredRecords; count += perHour {
		for inHour := 0; inHour < perHour; inHour++ {
			datumTime := startTime.Add(time.Duration(-(count + inHour + 1)) * time.Hour / time.Duration(perHour))

			datum := NewGlucose(&typ, &Units, &datumTime, &deviceId, &uploadId)
			datum.Value = pointer.FromAny(rand.NormFloat64()*StandardDeviation + VeryHighBloodGlucose)

			values = append(values, *datum.Value)

			dataSetData[requiredRecords-(count+inHour+1)] = datum
		}
	}

	return dataSetData, PopStdDev(values)
}

func CreateGlucoseBuckets(startTime time.Time, hours int, recordsPerBucket int, minutes bool) []*types.Bucket[*types.GlucoseBucket, types.GlucoseBucket] {
	buckets := make([]*types.Bucket[*types.GlucoseBucket, types.GlucoseBucket], hours)

	for i := 0; i < hours; i++ {
		buckets[i] = &types.Bucket[*types.GlucoseBucket, types.GlucoseBucket]{
			BucketShared: types.BucketShared{
				Type:      types.SummaryTypeCGM,
				Time:      startTime.Add(time.Hour * time.Duration(i)),
				FirstData: startTime.Add(time.Hour * time.Duration(i)),
				LastData:  startTime.Add(time.Hour*time.Duration(i) + time.Hour - 1*time.Millisecond),
			},
			Data: &types.GlucoseBucket{
				LastRecordDuration: 5,
			},
		}

		ranges := []*types.Range{
			&buckets[i].Data.VeryLow,
			&buckets[i].Data.Low,
			&buckets[i].Data.Target,
			&buckets[i].Data.High,
			&buckets[i].Data.VeryHigh,
			&buckets[i].Data.AnyLow,
			&buckets[i].Data.AnyHigh,
		}

		for k := range ranges {
			ranges[k].Records = recordsPerBucket
			ranges[k].Variance = 1

			if minutes {
				ranges[k].Minutes = recordsPerBucket * 5
			}
		}

		buckets[i].Data.Total.Glucose = float64(recordsPerBucket) * InTargetBloodGlucose * 5
		buckets[i].Data.Total.Records = recordsPerBucket
		buckets[i].Data.Total.Variance = 1

		if minutes {
			buckets[i].Data.Total.Minutes = recordsPerBucket * 5
		}
	}

	return buckets
}
