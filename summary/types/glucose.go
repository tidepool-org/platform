package types

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/blood/glucose"
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
)

const MaxRecordsPerBucket = 60 // one per minute max

type GlucosePeriods map[string]*GlucosePeriod

type Range struct {
	Glucose float64 `json:"glucose,omitempty" bson:"glucose,omitempty"`
	Minutes int     `json:"minutes,omitempty" bson:"minutes,omitempty"`
	Records int     `json:"records,omitempty" bson:"records,omitempty"`

	Percent  float64 `json:"percent,omitempty" bson:"percent,omitempty"`
	Variance float64 `json:"variance,omitempty" bson:"variance,omitempty"`
}

func (r *Range) Add(new *Range) {
	r.Variance = r.CombineVariance(new)
	r.Glucose += new.Glucose
	r.Minutes += new.Minutes
	r.Records += new.Records

	// clear percent, we don't have required values at this stage
	r.Percent = 0
}

func (r *Range) Update(record *glucoseDatum.Glucose) {
	r.Minutes += GetDuration(record)
	r.Records++
}

func (r *Range) UpdateTotal(record *glucoseDatum.Glucose) {
	normalizedValue := *glucose.NormalizeValueForUnits(record.Value, record.Units)

	// if this is bgm data, this will return 0
	duration := GetDuration(record)

	// this must occur before the regular update as the pre-increment counters are used during calc
	if duration > 0 {
		r.Variance = r.CalculateVariance(normalizedValue, float64(duration))
		r.Glucose += normalizedValue * float64(duration)
	} else {
		r.Glucose += normalizedValue
	}

	r.Update(record)
}

// CombineVariance Implemented using https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance#Parallel_algorithm
func (r *Range) CombineVariance(new *Range) float64 {
	// Exit early for No-Op case
	if r.Variance == 0 && new.Variance == 0 {
		return 0
	}

	// Return new if existing is 0
	if r.Variance == 0 {
		return new.Variance
	}

	// if we have no values in either bucket, this will result in NaN, and cant be added anyway, return what we started with
	if r.Minutes == 0 || new.Minutes == 0 {
		return r.Variance
	}

	n1 := float64(r.Minutes)
	n2 := float64(new.Minutes)
	n := n1 + n2
	delta := new.Glucose/n2 - r.Glucose/n1
	return r.Variance + new.Variance + math.Pow(delta, 2)*n1*n2/n
}

// CalculateVariance Implemented using https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance#Weighted_incremental_algorithm
func (r *Range) CalculateVariance(value float64, duration float64) float64 {
	var mean float64 = 0
	if r.Minutes > 0 {
		mean = r.Glucose / float64(r.Minutes)
	}

	weight := float64(r.Minutes) + duration
	newMean := mean + (duration/weight)*(value-mean)
	return r.Variance + duration*(value-mean)*(value-newMean)
}

func (r *Range) CalculateDelta(currentRange, offsetRange *Range) {
	r.Percent = currentRange.Percent - offsetRange.Percent
	r.Records = currentRange.Records - offsetRange.Records
	r.Minutes = currentRange.Minutes - offsetRange.Minutes
}

type GlucoseRanges struct {
	Total       Range `json:"total" bson:"total"`
	VeryLow     Range `json:"inVeryLow" bson:"inVeryLow"`
	Low         Range `json:"inLow" bson:"inLow"`
	Target      Range `json:"inTarget" bson:"inTarget"`
	High        Range `json:"inHigh" bson:"inHigh"`
	VeryHigh    Range `json:"inVeryHigh" bson:"inVeryHigh"`
	ExtremeHigh Range `json:"inExtremeHigh" bson:"inExtremeHigh"`
	AnyLow      Range `json:"inAnyLow" bson:"inAnyLow"`
	AnyHigh     Range `json:"inAnyHigh" bson:"inAnyHigh"`
}

func (rs *GlucoseRanges) Add(new *GlucoseRanges) {
	rs.Total.Add(&new.Total)
	rs.VeryLow.Add(&new.VeryLow)
	rs.Low.Add(&new.Low)
	rs.Target.Add(&new.Target)
	rs.High.Add(&new.High)
	rs.VeryHigh.Add(&new.VeryHigh)
	rs.ExtremeHigh.Add(&new.ExtremeHigh)
	rs.AnyLow.Add(&new.AnyLow)
	rs.AnyHigh.Add(&new.AnyHigh)
}

func (rs *GlucoseRanges) finalizeMinutes(days int) {
	rs.Total.Percent = float64(rs.Total.Minutes) / float64(days*24*60)
	rs.VeryLow.Percent = float64(rs.VeryLow.Minutes) / float64(rs.Total.Minutes)
	rs.Low.Percent = float64(rs.Low.Minutes) / float64(rs.Total.Minutes)
	rs.Target.Percent = float64(rs.Target.Minutes) / float64(rs.Total.Minutes)
	rs.High.Percent = float64(rs.High.Minutes) / float64(rs.Total.Minutes)
	rs.VeryHigh.Percent = float64(rs.VeryHigh.Minutes) / float64(rs.Total.Minutes)
	rs.ExtremeHigh.Percent = float64(rs.ExtremeHigh.Minutes) / float64(rs.Total.Minutes)
	rs.AnyLow.Percent = float64(rs.AnyLow.Minutes) / float64(rs.Total.Minutes)
	rs.AnyHigh.Percent = float64(rs.AnyHigh.Minutes) / float64(rs.Total.Minutes)
}

func (rs *GlucoseRanges) finalizeRecords() {
	// Total percent is only valid for minutes, when data represents real CGM time
	rs.Total.Percent = 0
	rs.VeryLow.Percent = float64(rs.VeryLow.Records) / float64(rs.Total.Records)
	rs.Low.Percent = float64(rs.Low.Records) / float64(rs.Total.Records)
	rs.Target.Percent = float64(rs.Target.Records) / float64(rs.Total.Records)
	rs.High.Percent = float64(rs.High.Records) / float64(rs.Total.Records)
	rs.VeryHigh.Percent = float64(rs.VeryHigh.Records) / float64(rs.Total.Records)
	rs.ExtremeHigh.Percent = float64(rs.ExtremeHigh.Records) / float64(rs.Total.Records)
	rs.AnyLow.Percent = float64(rs.AnyLow.Records) / float64(rs.Total.Records)
	rs.AnyHigh.Percent = float64(rs.AnyHigh.Records) / float64(rs.Total.Records)
}

func (rs *GlucoseRanges) Finalize(days int) {
	if rs.Total.Minutes != 0 {
		// if our bucket (period, at this point) has minutes
		rs.finalizeMinutes(days)
	} else if rs.Total.Records != 0 {
		// otherwise, we only have record counts
		rs.finalizeRecords()
	}
}

func (rs *GlucoseRanges) Update(record *glucoseDatum.Glucose) {
	normalizedValue := *glucose.NormalizeValueForUnits(record.Value, record.Units)

	if normalizedValue < veryLowBloodGlucose {
		rs.VeryLow.Update(record)
		rs.AnyLow.Update(record)
	} else if normalizedValue > veryHighBloodGlucose {
		rs.VeryHigh.Update(record)
		rs.AnyHigh.Update(record)

		// VeryHigh is inclusive of extreme high, this is intentional
		if normalizedValue >= extremeHighBloodGlucose {
			rs.ExtremeHigh.Update(record)
		}
	} else if normalizedValue < lowBloodGlucose {
		rs.Low.Update(record)
		rs.AnyLow.Update(record)
	} else if normalizedValue > highBloodGlucose {
		rs.AnyHigh.Update(record)
		rs.High.Update(record)
	} else {
		rs.Target.Update(record)
	}

	rs.Total.UpdateTotal(record)
}

func (rs *GlucoseRanges) CalculateDelta(current, previous *GlucoseRanges) {
	rs.Total.CalculateDelta(&current.Total, &previous.Total)
	rs.VeryLow.CalculateDelta(&current.VeryLow, &previous.VeryLow)
	rs.Low.CalculateDelta(&current.Low, &previous.Low)
	rs.Target.CalculateDelta(&current.Target, &previous.Target)
	rs.High.CalculateDelta(&current.High, &previous.High)
	rs.VeryHigh.CalculateDelta(&current.VeryHigh, &previous.VeryHigh)
	rs.ExtremeHigh.CalculateDelta(&current.ExtremeHigh, &previous.ExtremeHigh)
	rs.AnyLow.CalculateDelta(&current.AnyLow, &previous.AnyLow)
	rs.AnyHigh.CalculateDelta(&current.AnyHigh, &previous.AnyHigh)
}

type GlucoseBucket struct {
	GlucoseRanges      `json:",inline" bson:",inline"`
	LastRecordDuration int `json:"lastRecordDuration" bson:"lastRecordDuration"`
}

func (b *GlucoseBucket) ShouldSkipDatum(d *glucoseDatum.Glucose, lastData *time.Time) bool {
	// if we have more records than could possibly be in 1 hour of data
	if b.Total.Records > MaxRecordsPerBucket {
		return true
	}

	// if we have cgm data, we care about blackout periods
	if d.Type == continuous.Type {
		// calculate blackoutWindow based on duration of previous value
		// remove 10 seconds from the duration to prevent slight early reporting or exactly on time reporting from being skipped.
		blackoutWindow := time.Duration(b.LastRecordDuration)*time.Minute - 10*time.Second

		// Skip record if we are within the blackout window
		if d.Time.Sub(*lastData) < blackoutWindow {
			return true
		}
	}

	return false
}

func (b *GlucoseBucket) Update(r data.Datum, lastData *time.Time) (bool, error) {
	record, ok := r.(*glucoseDatum.Glucose)
	if !ok {
		return false, errors.New("record for calculation is not compatible with Glucose type")
	}

	if b.ShouldSkipDatum(record, lastData) {
		return false, nil
	}

	b.GlucoseRanges.Update(record)
	b.LastRecordDuration = GetDuration(record)

	return true, nil
}

type GlucosePeriod struct {
	GlucoseRanges `json:",inline" bson:",inline"`
	HoursWithData int `json:"hoursWithData,omitempty" bson:"hoursWithData,omitempty"`
	DaysWithData  int `json:"daysWithData,omitempty" bson:"daysWithData,omitempty"`

	AverageGlucose             float64 `json:"averageGlucoseMmol,omitempty" bson:"averageGlucoseMmol,omitempty"`
	GlucoseManagementIndicator float64 `json:"glucoseManagementIndicator,omitempty" bson:"glucoseManagementIndicator,omitempty"`

	CoefficientOfVariation float64 `json:"coefficientOfVariation,omitempty" bson:"coefficientOfVariation,omitempty"`
	StandardDeviation      float64 `json:"standardDeviation,omitempty" bson:"standardDeviation,omitempty"`

	AverageDailyRecords float64 `json:"averageDailyRecords,omitempty" bson:"averageDailyRecords,omitempty,omitempty"`

	Delta *GlucosePeriod `json:"delta,omitempty" bson:"delta,omitempty"`

	state CalcState
}

func (p *GlucosePeriod) CalculateDelta(current *GlucosePeriod, previous *GlucosePeriod) {
	p.GlucoseRanges.CalculateDelta(&current.GlucoseRanges, &previous.GlucoseRanges)

	Delta(&current.AverageGlucose, &previous.AverageGlucose, &p.AverageGlucose)
	Delta(&current.GlucoseManagementIndicator, &previous.GlucoseManagementIndicator, &p.GlucoseManagementIndicator)
	Delta(&current.AverageDailyRecords, &previous.AverageDailyRecords, &p.AverageDailyRecords)
	Delta(&current.StandardDeviation, &previous.StandardDeviation, &p.StandardDeviation)
	Delta(&current.CoefficientOfVariation, &previous.CoefficientOfVariation, &p.CoefficientOfVariation)
	Delta(&current.DaysWithData, &previous.DaysWithData, &p.DaysWithData)
	Delta(&current.HoursWithData, &previous.HoursWithData, &p.HoursWithData)
}

func (p *GlucosePeriod) Update(bucket *Bucket[*GlucoseBucket, GlucoseBucket]) error {
	if p.state.Final {
		return errors.New("period has been finalized, cannot add any data")
	}

	if bucket.Data.Total.Records == 0 {
		return nil
	}

	if p.state.LastCountedDay.IsZero() {
		p.state.FirstCountedDay = bucket.Time
		p.state.LastCountedDay = bucket.Time

		p.state.FirstCountedHour = bucket.Time
		p.state.LastCountedHour = bucket.Time

		p.state.FirstData = bucket.FirstData
		p.state.LastData = bucket.LastData

		p.state.LastRecordDuration = bucket.Data.LastRecordDuration

		p.DaysWithData++
		p.HoursWithData++
	} else {
		if bucket.Time.Before(p.state.FirstCountedHour) {
			p.HoursWithData++
			p.state.FirstCountedHour = bucket.Time
			p.state.FirstData = bucket.FirstData

			if p.state.FirstCountedDay.Sub(bucket.Time).Hours() >= 24 {
				p.state.FirstCountedDay = bucket.Time
				p.DaysWithData++
			}
		} else if bucket.Time.After(p.state.LastCountedHour) {
			p.HoursWithData++
			p.state.LastCountedHour = bucket.Time
			p.state.LastData = bucket.LastData
			p.state.LastRecordDuration = bucket.Data.LastRecordDuration

			if bucket.Time.Sub(p.state.LastCountedDay).Hours() >= 24 {
				p.state.LastCountedDay = bucket.Time
				p.DaysWithData++
			}
		} else {
			return fmt.Errorf("bucket of time %s is within the existing period range of %s - %s",
				bucket.Time, p.state.FirstCountedHour, p.state.LastCountedHour)
		}
	}

	p.Add(&bucket.Data.GlucoseRanges)

	return nil
}

func (p *GlucosePeriod) Finalize(days int) {
	if p.state.Final != false {
		return
	}
	p.GlucoseRanges.Finalize(days)

	if p.Total.Glucose != 0 {
		if p.Total.Minutes != 0 {
			// if we have minutes
			p.AverageGlucose = p.Total.Glucose / float64(p.Total.Minutes)
			p.GlucoseManagementIndicator = CalculateGMI(p.AverageGlucose)
			p.StandardDeviation = math.Sqrt(p.Total.Variance / float64(p.Total.Minutes))
			p.CoefficientOfVariation = p.StandardDeviation / p.AverageGlucose
		} else if p.Total.Records != 0 {
			// if we have only records
			p.AverageGlucose = p.Total.Glucose / float64(p.Total.Records)
		}
	}

	if p.Total.Records != 0 {
		p.AverageDailyRecords = float64(p.Total.Records) / float64(days)
	}

	p.state.Final = true
}

func (st *GlucosePeriods) Init() {
	*st = make(GlucosePeriods)
}

func (st *GlucosePeriods) Update(ctx context.Context, bucketsCursor *mongo.Cursor) error {
	// count backwards (newest first) through hourly stats, stopping at 1d, 7d, 14d, 30d
	period := GlucosePeriod{}
	offsetPeriod := GlucosePeriod{}
	offsetPeriods := make(GlucosePeriods)

	var stopPoints []time.Time
	nextStopPoint := 0

	var offsetStopPoints []time.Time
	nextOffsetStopPoint := 0

	previousBucketTime := time.Time{}

	for bucketsCursor.Next(ctx) {
		bucket := &Bucket[*GlucoseBucket, GlucoseBucket]{}
		if err := bucketsCursor.Decode(bucket); err != nil {
			return err
		}

		if !previousBucketTime.IsZero() && bucket.Time.Compare(previousBucketTime) >= 0 {
			return fmt.Errorf("bucket with date %s is equal or later than the last added bucket with date %s, "+
				"buckets must be in reverse order and unique", bucket.Time, previousBucketTime)
		}
		previousBucketTime = bucket.Time

		// Use the newest (last) bucket here to calculate date ranges
		if stopPoints == nil {
			stopPoints, offsetStopPoints = calculateStopPoints(bucket.Time)
		}

		if bucket.Data.Total.Records == 0 {
			panic("bucket exists with 0 records")
		}

		if len(stopPoints) > nextStopPoint && bucket.Time.Compare(stopPoints[nextStopPoint]) <= 0 {
			st.CalculatePeriod(periodLengths[nextStopPoint], period)
			nextStopPoint++
		}

		if len(offsetStopPoints) > nextOffsetStopPoint && bucket.Time.Compare(offsetStopPoints[nextOffsetStopPoint]) <= 0 {
			CalculateOffsetPeriod(offsetPeriods, periodLengths[nextOffsetStopPoint], offsetPeriod)
			offsetPeriod = GlucosePeriod{}
			nextOffsetStopPoint++
		}

		// only count primary stats when the next stop point is a real period
		if len(stopPoints) > nextStopPoint {
			if err := period.Update(bucket); err != nil {
				return err
			}
		}

		// only add to offset stats when primary stop point is ahead of offset
		if nextStopPoint > nextOffsetStopPoint && len(offsetStopPoints) > nextOffsetStopPoint {
			if err := offsetPeriod.Update(bucket); err != nil {
				return err
			}
		}
	}

	// fill in periods we never reached
	for i := nextStopPoint; i < len(stopPoints); i++ {
		st.CalculatePeriod(periodLengths[i], period)
	}
	for i := nextOffsetStopPoint; i < len(offsetStopPoints); i++ {
		CalculateOffsetPeriod(offsetPeriods, periodLengths[i], offsetPeriod)
		offsetPeriod = GlucosePeriod{}
	}

	st.CalculateDelta(offsetPeriods)

	return nil
}

func (st *GlucosePeriods) CalculateDelta(offsetPeriods GlucosePeriods) {
	for k := range *st {
		d := &GlucosePeriod{}
		d.CalculateDelta((*st)[k], offsetPeriods[k])
		(*st)[k].Delta = d
	}
}

func (st *GlucosePeriods) CalculatePeriod(days int, period GlucosePeriod) {
	period.Finalize(days)
	(*st)[strconv.Itoa(days)+"d"] = &period
}

func CalculateOffsetPeriod(offsetPeriods GlucosePeriods, days int, period GlucosePeriod) {
	period.Finalize(days)
	offsetPeriods[strconv.Itoa(days)+"d"] = &period
}
