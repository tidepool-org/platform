package test

import (
	"math"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/data/test"
	baseDatum "github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/data/types/upload"
	dataTypesUploadTest "github.com/tidepool-org/platform/data/types/upload/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/summary/types"
)

const (
	VeryLowBloodGlucose     = 3.0
	LowBloodGlucose         = 3.9
	HighBloodGlucose        = 10.0
	VeryHighBloodGlucose    = 13.9
	ExtremeHighBloodGlucose = 19.4
	InTargetBloodGlucose    = 5.0
)

func SliceToCursor[T any](s []T) (*mongo.Cursor, error) {
	return mongo.NewCursorFromDocuments(ConvertToIntArray(s), nil, nil)
}

func ConvertToIntArray[T any](arr []T) []interface{} {
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

func Mean(x []float64) float64 {
	sum := 0.0
	for i := 0; i < len(x); i++ {
		sum += x[i]
	}
	return sum / float64(len(x))
}

func CalculateVariance(x []float64, mean float64) float64 {
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

func CalculateStdDevAndVariance(x []float64) (float64, float64) {
	variance := CalculateVariance(x, Mean(x)) / float64(len(x))
	return math.Sqrt(variance), variance
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

func CreateGlucoseBuckets(startTime time.Time, hours int, recordsPerBucket int, minutes bool) []*types.Bucket[*types.GlucoseBucket, types.GlucoseBucket] {
	buckets := make([]*types.Bucket[*types.GlucoseBucket, types.GlucoseBucket], hours)

	startTime = startTime.Add(time.Hour * time.Duration(hours))

	for i := 0; i < hours; i++ {
		buckets[i] = &types.Bucket[*types.GlucoseBucket, types.GlucoseBucket]{
			BaseBucket: types.BaseBucket{
				Type:      types.SummaryTypeCGM,
				Time:      startTime.Add(-time.Hour * time.Duration(i)),
				FirstData: startTime.Add(-time.Hour * time.Duration(i)),
				LastData:  startTime.Add(-time.Hour*time.Duration(i) + time.Hour - 5*time.Minute),
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
			&buckets[i].Data.ExtremeHigh,
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

		glucoseMultiplier := 1.0
		if minutes {
			glucoseMultiplier = 5.0
			buckets[i].Data.Total.Minutes = recordsPerBucket * 5
		}

		buckets[i].Data.Total.Glucose = float64(recordsPerBucket) * InTargetBloodGlucose * glucoseMultiplier
		buckets[i].Data.Total.Records = recordsPerBucket
		buckets[i].Data.Total.Variance = 1

	}

	return buckets
}

func CreateContinuousBuckets(startTime time.Time, hours int, recordsPerBucket int) []*types.Bucket[*types.ContinuousBucket, types.ContinuousBucket] {
	buckets := make([]*types.Bucket[*types.ContinuousBucket, types.ContinuousBucket], hours)

	startTime = startTime.Add(time.Hour * time.Duration(hours))

	for i := 0; i < hours; i++ {
		buckets[i] = &types.Bucket[*types.ContinuousBucket, types.ContinuousBucket]{
			BaseBucket: types.BaseBucket{
				Type:      types.SummaryTypeCGM,
				Time:      startTime.Add(-time.Hour * time.Duration(i)),
				FirstData: startTime.Add(-time.Hour * time.Duration(i)),
				LastData:  startTime.Add(-time.Hour*time.Duration(i) + time.Hour - 5*time.Minute),
			},
			Data: &types.ContinuousBucket{},
		}

		ranges := []*types.Range{
			&buckets[i].Data.Realtime,
			&buckets[i].Data.Deferred,
		}

		for k := range ranges {
			ranges[k].Records = recordsPerBucket
		}

		buckets[i].Data.Total.Records = recordsPerBucket
	}

	return buckets
}

func NewDeferredGlucose(typ string, datumTime time.Time, value float64) (g *glucose.Glucose) {
	g = NewGlucose(&typ, &Units, &datumTime, pointer.FromAny("SummaryTestDevice"), pointer.FromAny(test.RandomSetID()))
	g.CreatedTime = pointer.FromAny(datumTime.AddDate(0, 0, 1))
	g.Value = &value
	return g
}

func NewRealtimeGlucose(typ string, datumTime time.Time, value float64) (g *glucose.Glucose) {
	g = NewGlucose(&typ, &Units, &datumTime, pointer.FromAny("SummaryTestDevice"), pointer.FromAny(test.RandomSetID()))
	g.CreatedTime = pointer.FromAny(datumTime.Add(5 * time.Minute))
	g.Value = &value
	return g
}

func NewDataSetDataRealtime(typ string, userId string, uploadId string, startTime time.Time, hours float64, realtime bool) []mongo.WriteModel {
	requiredRecords := int(hours * 2)
	dataSetData := make([]mongo.WriteModel, requiredRecords)
	deviceId := "SummaryTestDevice"

	glucoseValue := pointer.FromAny(InTargetBloodGlucose)

	// generate X hours of data
	for count := 0; count < requiredRecords; count += 1 {
		datumTime := startTime.Add(time.Duration(count-requiredRecords) * time.Minute * 30)

		datum := NewGlucose(&typ, &Units, &datumTime, &deviceId, &uploadId)
		datum.Value = glucoseValue
		datum.UserID = &userId

		if realtime {
			datum.CreatedTime = pointer.FromAny(datumTime.Add(5 * time.Minute))
		}

		dataSetData[count] = mongo.NewInsertOneModel().SetDocument(datum)
	}

	return dataSetData
}

func SliceToInsertWriteModel[T any](d []T) []mongo.WriteModel {
	w := make([]mongo.WriteModel, len(d))

	for i := 0; i < len(d); i++ {
		w[i] = mongo.NewInsertOneModel().SetDocument(d[i])
	}

	return w
}

func NewDataSet(userID string, typ string) *upload.Upload {
	var deviceId = "SummaryTestDevice"
	var timestamp = time.Now().UTC().Truncate(time.Millisecond)

	dataSet := dataTypesUploadTest.RandomUpload()
	dataSet.DataSetType = &typ
	dataSet.Active = true
	dataSet.ArchivedDataSetID = nil
	dataSet.ArchivedTime = nil
	dataSet.CreatedTime = &timestamp
	dataSet.CreatedUserID = nil
	dataSet.DeletedTime = nil
	dataSet.DeletedUserID = nil
	dataSet.DeviceID = &deviceId
	dataSet.Location.GPS.Origin.Time = nil
	dataSet.ModifiedTime = &timestamp
	dataSet.ModifiedUserID = nil
	dataSet.Origin.Time = nil
	dataSet.UserID = &userID
	return dataSet
}

func NewDataSetData(typ string, userId string, startTime time.Time, hours float64, glucoseValue float64) []mongo.WriteModel {
	requiredRecords := int(hours * 1)
	var dataSetData = make([]mongo.WriteModel, requiredRecords)

	for count := 0; count < requiredRecords; count++ {
		datumTime := startTime.Add(time.Duration(-(count + 1)) * time.Minute * 60)
		datum := NewGlucoseWithValue(typ, datumTime, glucoseValue)
		datum.UserID = &userId
		dataSetData[count] = mongo.NewInsertOneModel().SetDocument(datum)
	}
	return dataSetData
}

func NewDataSetDataModifiedTime(typ string, userId string, startTime time.Time, modifiedTime time.Time, hours float64, glucoseValue float64) []mongo.WriteModel {
	requiredRecords := int(hours * 1)
	var dataSetData = make([]mongo.WriteModel, requiredRecords)

	for count := 0; count < requiredRecords; count++ {
		datumTime := startTime.Add(time.Duration(-(count + 1)) * time.Minute * 60)
		datum := NewGlucoseWithValue(typ, datumTime, glucoseValue)
		datum.UserID = &userId
		datum.ModifiedTime = &modifiedTime
		dataSetData[count] = mongo.NewInsertOneModel().SetDocument(datum)
	}
	return dataSetData
}

func NewDatum(typ string) *baseDatum.Base {
	datum := baseDatum.New(typ)
	datum.Time = pointer.FromAny(time.Now().UTC())
	datum.Active = true
	return &datum
}

func NewOldDatum(typ string) *baseDatum.Base {
	datum := NewDatum(typ)
	datum.Active = true
	datum.Time = pointer.FromAny(time.Now().UTC().AddDate(0, -24, -1))
	return datum
}

func NewNewDatum(typ string) *baseDatum.Base {
	datum := NewDatum(typ)
	datum.Active = true
	datum.Time = pointer.FromAny(time.Now().UTC().AddDate(0, 0, 2))
	return datum
}
