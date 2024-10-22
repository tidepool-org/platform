package types

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/summary/fetcher"
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
)

type GlucosePeriods map[string]*GlucosePeriod

type GlucoseStats struct {
	Periods       GlucosePeriods `json:"periods,omitempty" bson:"periods,omitempty"`
	OffsetPeriods GlucosePeriods `json:"offsetPeriods,omitempty" bson:"offsetPeriods,omitempty"`
	TotalHours    int            `json:"totalHours,omitempty" bson:"totalHours,omitempty"`
}

type Range struct {
	Glucose float64 `json:"glucose,omitempty" bson:"glucose,omitempty"`
	Minutes int     `json:"minutes,omitempty" bson:"minutes,omitempty"`
	Records int     `json:"records,omitempty" bson:"records,omitempty"`

	Percent  float64 `json:"percent,omitempty" bson:"percent,omitempty"`
	Variance float64 `json:"variance,omitempty" bson:"variance,omitempty"`
}

func (R *Range) Add(new *Range) {
	R.Variance = R.CombineVariance(new)
	R.Glucose += new.Glucose
	R.Minutes += new.Minutes
	R.Records += new.Records

	// skip percent, they cannot be added here
}

func (R *Range) Update(value float64, duration int, total bool) {
	// this must occur before the counters below as the pre-increment counters are used during calc
	if total {
		R.Variance = R.CalculateVariance(value, float64(duration))
		R.Glucose += value * float64(duration)
	}

	R.Minutes += duration
	R.Records++
}

// CombineVariance Implemented using https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance#Parallel_algorithm
func (R *Range) CombineVariance(new *Range) float64 {
	// Exit early for No-Op case
	if R.Variance == 0 && new.Variance == 0 {
		return 0
	}

	// Return new if existing is 0
	if R.Variance == 0 {
		return new.Variance
	}

	// if we have no values in either bucket, this will result in NaN, and cant be added anyway, return what we started with
	if R.Minutes == 0 || new.Minutes == 0 {
		return R.Variance
	}

	n1 := float64(R.Minutes)
	n2 := float64(new.Minutes)
	n := n1 + n2
	delta := new.Glucose/n2 - R.Glucose/n1
	return R.Variance + new.Variance + math.Pow(delta, 2)*n1*n2/n
}

// CalculateVariance Implemented using https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance#Weighted_incremental_algorithm
func (R *Range) CalculateVariance(value float64, duration float64) float64 {
	var mean float64 = 0
	if R.Minutes > 0 {
		mean = R.Glucose / float64(R.Minutes)
	}

	weight := float64(R.Minutes) + duration
	newMean := mean + (duration/weight)*(value-mean)
	return R.Variance + duration*(value-mean)*(value-newMean)
}

type GlucoseRanges struct {
	Total       Range `json:"cgmUse,omitempty" bson:"cgmUse,omitempty"`
	VeryLow     Range `json:"inVeryLow,omitempty" bson:"inVeryLow,omitempty"`
	Low         Range `json:"inLow,omitempty" bson:"inLow,omitempty"`
	Target      Range `json:"inTarget,omitempty" bson:"inTarget,omitempty"`
	High        Range `json:"inHigh,omitempty" bson:"inHigh,omitempty"`
	VeryHigh    Range `json:"inVeryHigh,omitempty" bson:"inVeryHigh,omitempty"`
	ExtremeHigh Range `json:"inExtremeHigh,omitempty" bson:"inExtremeHigh,omitempty"`
	AnyLow      Range `json:"inAnyLow,omitempty" bson:"inAnyLow,omitempty"`
	AnyHigh     Range `json:"inAnyHigh,omitempty" bson:"inAnyHigh,omitempty"`
}

func (R *GlucoseRanges) Add(new *GlucoseRanges) {
	R.Total.Add(&new.Total)
	R.VeryLow.Add(&new.VeryLow)
	R.Low.Add(&new.Low)
	R.Target.Add(&new.Target)
	R.High.Add(&new.High)
	R.VeryHigh.Add(&new.VeryHigh)
	R.ExtremeHigh.Add(&new.ExtremeHigh)
	R.AnyLow.Add(&new.AnyLow)
	R.AnyHigh.Add(&new.AnyHigh)
}

type GlucoseBucket struct {
	GlucoseRanges
	LastRecordDuration int `json:"lastRecordDuration,omitempty" bson:"lastRecordDuration,omitempty"`
}

func (R *GlucoseRanges) finalizeMinutes(wallMinutes float64, days int) {
	R.Total.Percent = float64(R.Total.Minutes) / float64(days*24*60)
	fmt.Println(R.Total.Percent, days, R.Total.Minutes)
	fmt.Println("if ", wallMinutes, " <= ", minutesPerDay, " && ", R.Total.Percent, " > ", 0.7, " ) || (", wallMinutes, " > ", minutesPerDay, " && ", R.Total.Minutes, " > ", minutesPerDay, " )")
	if (wallMinutes <= minutesPerDay && R.Total.Percent > 0.7) || (wallMinutes > minutesPerDay && R.Total.Minutes > minutesPerDay) {
		fmt.Println("calculating percentages")
		R.VeryLow.Percent = float64(R.VeryLow.Minutes) / wallMinutes
		R.Low.Percent = float64(R.Low.Minutes) / wallMinutes
		R.Target.Percent = float64(R.Target.Minutes) / wallMinutes
		R.High.Percent = float64(R.High.Minutes) / wallMinutes
		R.VeryHigh.Percent = float64(R.VeryHigh.Minutes) / wallMinutes
		fmt.Println(R.VeryHigh.Percent, R.VeryHigh.Minutes, wallMinutes)
		R.ExtremeHigh.Percent = float64(R.ExtremeHigh.Minutes) / wallMinutes
		R.AnyLow.Percent = float64(R.AnyLow.Minutes) / wallMinutes
		R.AnyHigh.Percent = float64(R.AnyHigh.Minutes) / wallMinutes
	}
}

func (R *GlucoseRanges) finalizeRecords() {
	R.VeryLow.Percent = float64(R.VeryLow.Records) / float64(R.Total.Records)
	R.Low.Percent = float64(R.Low.Records) / float64(R.Total.Records)
	R.Target.Percent = float64(R.Target.Records) / float64(R.Total.Records)
	R.High.Percent = float64(R.High.Records) / float64(R.Total.Records)
	R.VeryHigh.Percent = float64(R.VeryHigh.Records) / float64(R.Total.Records)
	R.ExtremeHigh.Percent = float64(R.ExtremeHigh.Records) / float64(R.Total.Records)
	R.AnyLow.Percent = float64(R.AnyLow.Records) / float64(R.Total.Records)
	R.AnyHigh.Percent = float64(R.AnyHigh.Records) / float64(R.Total.Records)
}

func (R *GlucoseRanges) Finalize(firstData, lastData time.Time, lastDuration int, days int) {
	if R.Total.Minutes != 0 {
		// if our bucket (period, at this point) has minutes
		wallMinutes := lastData.Sub(firstData).Minutes() + float64(lastDuration)
		R.finalizeMinutes(wallMinutes, days)
	} else {
		// otherwise, we only have record counts
		R.finalizeRecords()
	}
}

func (R *GlucoseRanges) Update(record glucoseDatum.Glucose, duration int) {
	normalizedValue := *glucose.NormalizeValueForUnits(record.Value, record.Units)

	if normalizedValue < veryLowBloodGlucose {
		R.VeryLow.Update(normalizedValue, duration, false)
		R.AnyLow.Update(normalizedValue, duration, false)
	} else if normalizedValue > veryHighBloodGlucose {
		R.VeryHigh.Update(normalizedValue, duration, false)
		R.AnyHigh.Update(normalizedValue, duration, false)

		// VeryHigh is inclusive of extreme high, this is intentional
		if normalizedValue >= extremeHighBloodGlucose {
			R.ExtremeHigh.Update(normalizedValue, duration, false)
		}
	} else if normalizedValue < lowBloodGlucose {
		R.Low.Update(normalizedValue, duration, false)
		R.AnyLow.Update(normalizedValue, duration, false)
	} else if normalizedValue > highBloodGlucose {
		R.AnyHigh.Update(normalizedValue, duration, false)
		R.High.Update(normalizedValue, duration, false)
	} else {
		R.Target.Update(normalizedValue, duration, false)
	}

	R.Total.Update(normalizedValue, duration, true)
}

func (B *GlucoseBucket) Add(bucket *GlucoseBucket) {
	B.Add(bucket)
}

func (B *GlucoseBucket) Update(r data.Datum, shared *BucketShared) error {
	record, ok := r.(*glucoseDatum.Glucose)
	if !ok {
		return errors.New("record for calculation is not compatible with Glucose type")
	}

	if DeviceDataToSummaryTypes[record.Type] != shared.Type {
		return fmt.Errorf("record for %s calculation is of invald type %s", shared.Type, record.Type)
	}

	// if this is bgm data, this will return 1
	duration := GetDuration(record)

	// if we have cgm data, we care about blackout periods
	if shared.Type == SummaryTypeContinuous {
		// calculate blackoutWindow based on duration of previous value
		blackoutWindow := time.Duration(B.LastRecordDuration)*time.Minute - 10*time.Second

		// Skip record if we are within the blackout window
		if record.Time.Sub(shared.LastData) < blackoutWindow {
			return nil
		}
	}

	B.GlucoseRanges.Update(*record, duration)

	B.LastRecordDuration = duration

	return nil
}

type GlucosePeriod struct {
	GlucoseRanges `json:",inline" bson:",inline"`
	HoursWithData int `json:"hoursWithData,omitempty" bson:"hoursWithData,omitempty"`
	DaysWithData  int `json:"daysWithData,omitempty" bson:"daysWithData,omitempty"`

	final bool

	firstCountedDay time.Time
	lastCountedDay  time.Time

	firstCountedHour time.Time
	lastCountedHour  time.Time

	lastData  time.Time
	firstData time.Time

	lastRecordDuration int

	AverageGlucose             float64 `json:"averageGlucoseMmol,omitempty" bson:"avgGlucose,omitempty"`
	GlucoseManagementIndicator float64 `json:"glucoseManagementIndicator,omitempty" bson:"GMI,omitempty"`

	CoefficientOfVariation float64 `json:"coefficientOfVariation,omitempty" bson:"CV,omitempty"`
	StandardDeviation      float64 `json:"standardDeviation,omitempty" bson:"SD,omitempty"`

	AverageDailyRecords float64 `json:"averageDailyRecords,omitempty" b;son:"avgDailyRecords,omitempty,omitempty"`

	Delta *GlucosePeriod `json:"delta,omitempty" bson:"delta,omitempty"`
}

func (P *GlucosePeriod) IsFinal() bool {
	return P.final
}

func (P *GlucosePeriod) Update(bucket *Bucket[*GlucoseBucket, GlucoseBucket]) error {
	if P.final {
		return errors.New("period has been finalized, cannot add any data")
	}

	if bucket.Data.Total.Records == 0 {
		return nil
	}

	// NOTE this works correctly for buckets in forward or backwards order, but not unordered, it must be added with consistent direction
	// TODO this could use some math with firstData/lastData to work with non-hourly buckets, but today they're hourly.
	// TODO should this be moved to a generic periods type as a Shared sidecar, days/hours is probably useful to other types
	// TODO average daily records could also be moved

	if P.lastCountedDay.IsZero() {
		P.firstCountedDay = bucket.Time
		P.lastCountedDay = bucket.Time

		P.firstCountedHour = bucket.Time
		P.lastCountedHour = bucket.Time

		P.firstData = bucket.FirstData
		P.lastData = bucket.LastData

		P.lastRecordDuration = bucket.Data.LastRecordDuration

		P.DaysWithData++
		P.HoursWithData++
	} else {
		if bucket.Time.Before(P.firstCountedHour) {
			P.HoursWithData++
			P.firstCountedHour = bucket.Time
			P.firstData = bucket.FirstData

			if P.firstCountedDay.Sub(bucket.Time).Hours() >= 24 {
				P.firstCountedDay = bucket.Time
				P.DaysWithData++
			}
		} else if bucket.Time.After(P.lastCountedHour) {
			P.HoursWithData++
			P.lastCountedHour = bucket.Time
			P.lastData = bucket.LastData
			P.lastRecordDuration = bucket.Data.LastRecordDuration

			if bucket.Time.Sub(P.lastCountedDay).Hours() >= 24 {
				P.lastCountedDay = bucket.Time
				P.DaysWithData++
			}
		} else {
			return fmt.Errorf("bucket of time %s is within the existing period range of %s - %s", bucket.Time, P.firstCountedHour, P.lastCountedHour)
		}
	}

	P.Add(&bucket.Data.GlucoseRanges)

	return nil
}

func (P *GlucosePeriod) Finalize(days int) {
	P.final = true
	P.GlucoseRanges.Finalize(P.firstData, P.lastData, P.lastRecordDuration, days)
	P.AverageGlucose = P.Total.Glucose / float64(P.Total.Minutes)

	// we only add GMI if cgm use >70%, otherwise clear it
	if P.Total.Percent > 0.7 {
		P.GlucoseManagementIndicator = CalculateGMI(P.AverageGlucose)
	}

	P.StandardDeviation = math.Sqrt(P.Total.Variance / float64(P.Total.Minutes))
	P.CoefficientOfVariation = P.StandardDeviation / P.AverageGlucose

	P.AverageDailyRecords = float64(P.Total.Records) / float64(days)
}

func (s *GlucoseStats) Init() {
	s.Periods = make(map[string]*GlucosePeriod)
	s.OffsetPeriods = make(map[string]*GlucosePeriod)
	s.TotalHours = 0
}

func (s *GlucoseStats) Update(ctx context.Context, shared SummaryShared, bucketsFetcher BucketFetcher[*GlucoseBucket, GlucoseBucket], cursor fetcher.DeviceDataCursor) error {
	// move all of this to a generic method? fetcher interface?

	hasMoreData := true
	var buckets BucketsByTime[*GlucoseBucket, GlucoseBucket]
	var err error
	var userData []data.Datum
	var startTime time.Time
	var endTime time.Time

	for hasMoreData {
		userData, err = cursor.GetNextBatch(ctx)
		if errors.Is(err, fetcher.ErrCursorExhausted) {
			hasMoreData = false
		} else if err != nil {
			return err
		}

		if len(userData) > 0 {
			startTime = userData[0].GetTime().UTC().Truncate(time.Hour)
			endTime = userData[len(userData)].GetTime().UTC().Truncate(time.Hour)
			buckets, err = bucketsFetcher.GetBuckets(ctx, shared.UserID, startTime, endTime)
			if err != nil {
				return err
			}

			err = buckets.Update(shared.UserID, shared.Type, userData)
			if err != nil {
				return err
			}

			err = bucketsFetcher.WriteModifiedBuckets(ctx, shared.Dates.LastUpdatedDate, buckets)
			if err != nil {
				return err
			}

		}
	}

	allBuckets, err := bucketsFetcher.GetAllBuckets(ctx, shared.UserID)
	if err != nil {
		return err
	}
	defer allBuckets.Close(ctx)

	err = s.CalculateSummary(ctx, allBuckets)
	if err != nil {
		return err
	}

	s.CalculateDelta()

	return nil
}

func (s *GlucoseStats) CalculateSummary(ctx context.Context, buckets fetcher.AnyCursor) error {
	// count backwards (newest first) through hourly stats, stopping at 1d, 7d, 14d, 30d,
	// currently only supports day precision
	nextStopPoint := 0
	nextOffsetStopPoint := 0
	totalStats := GlucosePeriod{}
	totalOffsetStats := GlucosePeriod{}
	var err error
	var stopPoints []time.Time

	bucket := &Bucket[*GlucoseBucket, GlucoseBucket]{}

	for buckets.Next(ctx) {
		if err = buckets.Decode(bucket); err != nil {
			return err
		}

		// We should have the newest (last) bucket here, use its date for breakpoints
		if stopPoints == nil {
			stopPoints = calculateStopPoints(bucket.Time)
		}

		if bucket.Data.Total.Records == 0 {
			panic("bucket exists with 0 records")
		}

		s.TotalHours++

		// only add to offset stats when primary stop point is ahead of offset
		if nextStopPoint > nextOffsetStopPoint && len(stopPoints) > nextOffsetStopPoint {
			err = totalOffsetStats.Update(bucket)
			if err != nil {
				return err
			}

			if bucket.Time.Before(stopPoints[nextOffsetStopPoint]) {
				s.CalculatePeriod(periodLengths[nextOffsetStopPoint], true, totalOffsetStats)
				nextOffsetStopPoint++
				totalOffsetStats = GlucosePeriod{}
			}
		}

		// only count primary stats when the next stop point is a real period
		if len(stopPoints) > nextStopPoint {
			err = totalStats.Update(bucket)
			if err != nil {
				return err
			}

			if bucket.Time.Before(stopPoints[nextStopPoint]) {
				s.CalculatePeriod(periodLengths[nextStopPoint], false, totalStats)
				nextStopPoint++
			}
		}
	}

	// fill in periods we never reached
	for i := nextStopPoint; i < len(stopPoints); i++ {
		s.CalculatePeriod(periodLengths[i], false, totalStats)
	}
	for i := nextOffsetStopPoint; i < len(stopPoints); i++ {
		s.CalculatePeriod(periodLengths[i], true, totalOffsetStats)
		totalOffsetStats = GlucosePeriod{}
	}

	return nil
}

func (s *GlucoseStats) CalculateDelta() {

	for k := range s.Periods {
		// initialize delta pointers, make sure we are starting from a clean delta period/no shared pointers
		s.Periods[k].Delta = &GlucosePeriod{}
		s.OffsetPeriods[k].Delta = &GlucosePeriod{}

		BinDelta(&s.Periods[k].Total, &s.OffsetPeriods[k].Total, &s.Periods[k].Delta.Total, &s.OffsetPeriods[k].Delta.Total)
		BinDelta(&s.Periods[k].VeryLow, &s.OffsetPeriods[k].VeryLow, &s.Periods[k].Delta.VeryLow, &s.OffsetPeriods[k].Delta.VeryLow)
		BinDelta(&s.Periods[k].Low, &s.OffsetPeriods[k].Low, &s.Periods[k].Delta.Low, &s.OffsetPeriods[k].Delta.Low)
		BinDelta(&s.Periods[k].Target, &s.OffsetPeriods[k].Target, &s.Periods[k].Delta.Target, &s.OffsetPeriods[k].Delta.Target)
		BinDelta(&s.Periods[k].High, &s.OffsetPeriods[k].High, &s.Periods[k].Delta.High, &s.OffsetPeriods[k].Delta.High)
		BinDelta(&s.Periods[k].VeryHigh, &s.OffsetPeriods[k].VeryHigh, &s.Periods[k].Delta.VeryHigh, &s.OffsetPeriods[k].Delta.VeryHigh)
		BinDelta(&s.Periods[k].ExtremeHigh, &s.OffsetPeriods[k].ExtremeHigh, &s.Periods[k].Delta.ExtremeHigh, &s.OffsetPeriods[k].Delta.ExtremeHigh)
		BinDelta(&s.Periods[k].AnyLow, &s.OffsetPeriods[k].AnyLow, &s.Periods[k].Delta.AnyLow, &s.OffsetPeriods[k].Delta.AnyLow)
		BinDelta(&s.Periods[k].AnyHigh, &s.OffsetPeriods[k].AnyHigh, &s.Periods[k].Delta.AnyHigh, &s.OffsetPeriods[k].Delta.AnyHigh)

		Delta(&s.Periods[k].AverageGlucose, &s.OffsetPeriods[k].AverageGlucose, &s.Periods[k].Delta.AverageGlucose, &s.OffsetPeriods[k].Delta.AverageGlucose)
		Delta(&s.Periods[k].GlucoseManagementIndicator, &s.OffsetPeriods[k].GlucoseManagementIndicator, &s.OffsetPeriods[k].Delta.GlucoseManagementIndicator, &s.Periods[k].Delta.GlucoseManagementIndicator)
		Delta(&s.Periods[k].AverageDailyRecords, &s.OffsetPeriods[k].AverageDailyRecords, &s.OffsetPeriods[k].Delta.AverageDailyRecords, &s.Periods[k].Delta.AverageDailyRecords)
		Delta(&s.Periods[k].StandardDeviation, &s.OffsetPeriods[k].StandardDeviation, &s.Periods[k].Delta.StandardDeviation, &s.OffsetPeriods[k].Delta.StandardDeviation)
		Delta(&s.Periods[k].CoefficientOfVariation, &s.OffsetPeriods[k].CoefficientOfVariation, &s.Periods[k].Delta.CoefficientOfVariation, &s.OffsetPeriods[k].Delta.CoefficientOfVariation)
		Delta(&s.Periods[k].DaysWithData, &s.OffsetPeriods[k].DaysWithData, &s.Periods[k].Delta.DaysWithData, &s.OffsetPeriods[k].Delta.DaysWithData)
		Delta(&s.Periods[k].HoursWithData, &s.OffsetPeriods[k].HoursWithData, &s.Periods[k].Delta.HoursWithData, &s.OffsetPeriods[k].Delta.HoursWithData)
	}
}

func (s *GlucoseStats) CalculatePeriod(i int, offset bool, period GlucosePeriod) {
	// We don't make a copy of period, as the struct has no pointers... right? you didn't add any pointers right?
	period.Finalize(i)

	if offset {
		s.OffsetPeriods[strconv.Itoa(i)+"d"] = &period
	} else {
		s.Periods[strconv.Itoa(i)+"d"] = &period
	}

}
