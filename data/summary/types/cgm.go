package types

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/data/summary/fetcher"

	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"

	"github.com/mitchellh/mapstructure"
	"github.com/tidepool-org/platform/data/blood/glucose"
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/pointer"
)

type GlucoseBucketData struct {
	LastRecordDuration int `json:"LastRecordDuration" bson:"LastRecordDuration,omitempty"`

	InTarget   *GlucoseBin `json:"inTarget" bson:"inTarget,omitempty"`
	InLow      *GlucoseBin `json:"inLow" bson:"inLow,omitempty"`
	InVeryLow  *GlucoseBin `json:"inVeryLow" bson:"inVeryLow,omitempty"`
	InHigh     *GlucoseBin `json:"inHigh" bson:"inHigh,omitempty"`
	InVeryHigh *GlucoseBin `json:"inVeryHigh" bson:"inVeryHigh,omitempty"`
	Total      *TotalBin   `json:"total" bson:"total,omitempty"`
}

type GlucoseBin struct {
	Percent float64 `json:"percent" bson:"percent,omitempty"`
	Minutes int     `json:"minutes" bson:"minutes,omitempty"`
	Records int     `json:"records" bson:"records,omitempty"`
}

type TotalBin struct {
	Glucose float64 `json:"glucose" bson:"glucose,omitempty"`
	Minutes int     `json:"minutes" bson:"minutes,omitempty"`
	Records int     `json:"records" bson:"records,omitempty"`
}

type GlucosePeriod struct {
	AverageGlucoseMmol         float64 `json:"averageGlucoseMmol" bson:"averageGlucoseMmol,omitempty" mapstructure:"_"`
	GlucoseManagementIndicator float64 `json:"glucoseManagementIndicator" bson:"glucoseManagementIndicator,omitempty" mapstructure:"_"`

	AverageDailyRecords float64 `json:"averageDailyRecords" bson:"averageDailyRecords,omitempty" mapstructure:"_"`

	Delta *GlucosePeriod `json:"delta" bson:"delta,omitempty" mapstructure:"_"`

	Total      *GlucoseBin `json:"cgmUse" bson:"cgmUse,omitempty"`
	InTarget   *GlucoseBin `json:"inTarget" bson:"inTarget,omitempty"`
	InLow      *GlucoseBin `json:"inLow" bson:"inLow,omitempty"`
	InVeryLow  *GlucoseBin `json:"inVeryLow" bson:"inVeryLow,omitempty"`
	InAnyLow   *GlucoseBin `json:"inAnyLow" bson:"inAnyLow,omitempty"`
	InHigh     *GlucoseBin `json:"inHigh" bson:"inHigh,omitempty"`
	InVeryHigh *GlucoseBin `json:"inVeryHigh" bson:"inVeryHigh,omitempty"`
	InAnyHigh  *GlucoseBin `json:"inAnyHigh" bson:"inAnyHigh,omitempty"`
}

type GlucosePeriods map[string]*GlucosePeriod

type CGMStats struct {
	Periods       GlucosePeriods                                   `json:"periods" bson:"periods"`
	OffsetPeriods GlucosePeriods                                   `json:"offsetPeriods" bson:"offsetPeriods"`
	Buckets       []*Bucket[*GlucoseBucketData, GlucoseBucketData] `json:"buckets" bson:"buckets"`
	TotalHours    int                                              `json:"totalHours" bson:"totalHours"`
}

func (*CGMStats) GetType() string {
	return SummaryTypeCGM
}

func (*CGMStats) GetDeviceDataTypes() []string {
	return []string{continuous.Type}
}

func (s *CGMStats) Init() {
	s.Buckets = make([]*Bucket[*GlucoseBucketData, GlucoseBucketData], 0)
	s.Periods = make(map[string]*GlucosePeriod)
	s.OffsetPeriods = make(map[string]*GlucosePeriod)
	s.TotalHours = 0
}

func (s *CGMStats) GetBucketsLen() int {
	return len(s.Buckets)
}

func (s *CGMStats) GetBucketDate(i int) time.Time {
	return s.Buckets[i].Date
}

func (s *CGMStats) ClearInvalidatedBuckets(earliestModified time.Time) (firstData time.Time) {
	if len(s.Buckets) == 0 {
		return
	} else if earliestModified.After(s.Buckets[len(s.Buckets)-1].LastRecordTime) {
		return s.Buckets[len(s.Buckets)-1].LastRecordTime
	} else if earliestModified.Before(s.Buckets[0].Date) || earliestModified.Equal(s.Buckets[0].Date) {
		// we are before all existing buckets, remake for GC
		s.Buckets = make([]*Bucket[*GlucoseBucketData, GlucoseBucketData], 0)
		return
	}

	offset := len(s.Buckets) - (int(s.Buckets[len(s.Buckets)-1].Date.Sub(earliestModified.UTC().Truncate(time.Hour)).Hours()) + 1)

	for i := offset; i < len(s.Buckets); i++ {
		s.Buckets[i] = nil
	}
	s.Buckets = s.Buckets[:offset]

	if len(s.Buckets) > 0 {
		return s.Buckets[len(s.Buckets)-1].LastRecordTime
	}
	return
}

func (s *CGMStats) Update(ctx context.Context, cursor fetcher.DeviceDataCursor) error {
	hasMoreData := true
	for hasMoreData {
		userData, err := cursor.GetNextBatch(ctx)
		if errors.Is(err, fetcher.ErrCursorExhausted) {
			hasMoreData = false
		} else if err != nil {
			return err
		}

		if len(userData) > 0 {
			err = AddData(&s.Buckets, userData)
			if err != nil {
				return err
			}
		}
	}

	s.CalculateSummary()

	return nil
}

func (B *GlucoseBucketData) CalculateStats(r any, lastRecordTime *time.Time) (bool, error) {
	dataRecord, ok := r.(*glucoseDatum.Glucose)
	if !ok {
		return false, errors.New("CGM record for calculation is not compatible with Glucose type")
	}

	// this is a new bucket, use current record as duration reference
	if B.LastRecordDuration == 0 {
		B.LastRecordDuration = GetDuration(dataRecord)
	}

	// calculate blackoutWindow based on duration of previous value
	blackoutWindow := time.Duration(B.LastRecordDuration)*time.Minute - 10*time.Second

	// Skip record unless we are beyond the blackout window
	if dataRecord.Time.Sub(*lastRecordTime) > blackoutWindow {
		normalizedValue := *glucose.NormalizeValueForUnits(dataRecord.Value, pointer.FromAny(glucose.MmolL))
		duration := GetDuration(dataRecord)

		if normalizedValue < veryLowBloodGlucose {
			B.InVeryLow.Minutes += duration
			B.InVeryLow.Records++
		} else if normalizedValue > veryHighBloodGlucose {
			B.InVeryHigh.Minutes += duration
			B.InVeryHigh.Records++
		} else if normalizedValue < lowBloodGlucose {
			B.InLow.Minutes += duration
			B.InLow.Records++
		} else if normalizedValue > highBloodGlucose {
			B.InHigh.Minutes += duration
			B.InHigh.Records++
		} else {
			B.InTarget.Minutes += duration
			B.InTarget.Records++
		}

		B.Total.Minutes += duration
		B.Total.Records++
		B.Total.Glucose += normalizedValue * float64(duration)
		B.LastRecordDuration = duration

		return false, nil
	}

	return true, nil
}

func (s *CGMStats) CalculateSummary() {
	// count backwards (newest first) through hourly stats, stopping at 24, 24*7, 24*14, 24*30
	// currently only supports day precision
	nextStopPoint := 0
	nextOffsetStopPoint := 0
	totalStats := &GlucoseBucketData{}
	totalOffsetStats := &GlucoseBucketData{}

	for i := 0; i < len(s.Buckets); i++ {
		currentIndex := len(s.Buckets) - 1 - i

		// only count primary stats when the next stop point is a real period
		if len(stopPoints) > nextStopPoint {
			if i == stopPoints[nextStopPoint]*24 {
				s.CalculatePeriod(stopPoints[nextStopPoint], false, totalStats)
				nextStopPoint++
			}

			totalStats.InTarget.Minutes += s.Buckets[currentIndex].Data.InTarget.Minutes
			totalStats.InTarget.Records += s.Buckets[currentIndex].Data.InTarget.Records

			totalStats.InLow.Minutes += s.Buckets[currentIndex].Data.InLow.Minutes
			totalStats.InLow.Records += s.Buckets[currentIndex].Data.InLow.Records

			totalStats.InVeryLow.Minutes += s.Buckets[currentIndex].Data.InVeryLow.Minutes
			totalStats.InVeryLow.Records += s.Buckets[currentIndex].Data.InVeryLow.Records

			totalStats.InHigh.Minutes += s.Buckets[currentIndex].Data.InHigh.Minutes
			totalStats.InHigh.Records += s.Buckets[currentIndex].Data.InHigh.Records

			totalStats.InVeryHigh.Minutes += s.Buckets[currentIndex].Data.InVeryHigh.Minutes
			totalStats.InVeryHigh.Records += s.Buckets[currentIndex].Data.InVeryHigh.Records

			totalStats.Total.Glucose += s.Buckets[currentIndex].Data.Total.Glucose
			totalStats.Total.Minutes += s.Buckets[currentIndex].Data.Total.Minutes
			totalStats.Total.Records += s.Buckets[currentIndex].Data.Total.Records
		}

		// only add to offset stats when primary stop point is ahead of offset
		if nextStopPoint > nextOffsetStopPoint && len(stopPoints) > nextOffsetStopPoint {
			if i == stopPoints[nextOffsetStopPoint]*24*2 {
				s.CalculatePeriod(stopPoints[nextOffsetStopPoint], true, totalOffsetStats)
				nextOffsetStopPoint++
				totalOffsetStats = &GlucoseBucketData{}
			}
			totalOffsetStats.InTarget.Minutes += s.Buckets[currentIndex].Data.InTarget.Minutes
			totalOffsetStats.InTarget.Records += s.Buckets[currentIndex].Data.InTarget.Records

			totalOffsetStats.InLow.Minutes += s.Buckets[currentIndex].Data.InLow.Minutes
			totalOffsetStats.InLow.Records += s.Buckets[currentIndex].Data.InLow.Records

			totalOffsetStats.InVeryLow.Minutes += s.Buckets[currentIndex].Data.InVeryLow.Minutes
			totalOffsetStats.InVeryLow.Records += s.Buckets[currentIndex].Data.InVeryLow.Records

			totalOffsetStats.InHigh.Minutes += s.Buckets[currentIndex].Data.InHigh.Minutes
			totalOffsetStats.InHigh.Records += s.Buckets[currentIndex].Data.InHigh.Records

			totalOffsetStats.InVeryHigh.Minutes += s.Buckets[currentIndex].Data.InVeryHigh.Minutes
			totalOffsetStats.InVeryHigh.Records += s.Buckets[currentIndex].Data.InVeryHigh.Records

			totalOffsetStats.Total.Glucose += s.Buckets[currentIndex].Data.Total.Glucose
			totalOffsetStats.Total.Minutes += s.Buckets[currentIndex].Data.Total.Minutes
			totalOffsetStats.Total.Records += s.Buckets[currentIndex].Data.Total.Records
		}
	}

	// fill in periods we never reached
	for i := nextStopPoint; i < len(stopPoints); i++ {
		s.CalculatePeriod(stopPoints[i], false, totalStats)
	}
	for i := nextOffsetStopPoint; i < len(stopPoints); i++ {
		s.CalculatePeriod(stopPoints[i], true, totalOffsetStats)
		totalOffsetStats = &GlucoseBucketData{}
	}

	s.TotalHours = len(s.Buckets)

	s.CalculateDelta()
}

func (s *CGMStats) CalculateDelta() {
	// We do this as a separate pass through the periods as the amount of tracking required to reverse the iteration
	// and fill this in during the period calculation would likely nullify any benefits, at least with the current
	// approach.

	for k := range s.Periods {
		periodMap := map[string]interface{}{}
		err := mapstructure.Decode(s.Periods[k], &periodMap)
		if err != nil {
			panic(err)
		}

		offsetPeriodMap := map[string]interface{}{}
		err = mapstructure.Decode(s.OffsetPeriods[k], &offsetPeriodMap)
		if err != nil {
			panic(err)
		}

		deltaMap := map[string]interface{}{}
		offsetDeltaMap := map[string]interface{}{}

		for key := range periodMap {
			for t := range periodMap[key].(map[string]interface{}) {
				switch periodMap[key].(map[string]interface{})[t].(type) {
				case int:
					delta := periodMap[key].(map[string]interface{})[t].(int) - offsetPeriodMap[key].(map[string]interface{})[t].(int)
					deltaMap[key].(map[string]interface{})[t] = delta
					offsetDeltaMap[key].(map[string]interface{})[t] = -delta

				case float64:
					delta := periodMap[key].(map[string]interface{})[t].(float64) - offsetPeriodMap[key].(map[string]interface{})[t].(float64)
					deltaMap[key].(map[string]interface{})[t] = delta
					offsetDeltaMap[key].(map[string]interface{})[t] = -delta
				}
			}
		}

		err = mapstructure.Decode(periodMap, s.Periods[k].Delta)
		if err != nil {
			panic(err)
		}

		err = mapstructure.Decode(offsetPeriodMap, s.OffsetPeriods[k])
		if err != nil {
			panic(err)
		}

		// TODO mapstructure ignored keys
	}
}

func (s *CGMStats) CalculatePeriod(i int, offset bool, totalStats *GlucoseBucketData) {
	periodMap := map[string]map[string]interface{}{}
	totalStatsMap := map[string]interface{}{}

	err := mapstructure.Decode(totalStats, &totalStatsMap)
	if err != nil {
		panic(err)
	}

	for key := range totalStatsMap {
		for t := range totalStatsMap[key].(map[string]interface{}) {
			periodMap[key][t] = totalStatsMap[key].(map[string]interface{})[t]
		}
	}

	var activePeriod *GlucosePeriod
	if offset {
		activePeriod = s.OffsetPeriods[strconv.Itoa(i)+"d"]
	} else {
		activePeriod = s.Periods[strconv.Itoa(i)+"d"]
	}

	err = mapstructure.Decode(periodMap, activePeriod)
	if err != nil {
		panic(err)
	}
}
