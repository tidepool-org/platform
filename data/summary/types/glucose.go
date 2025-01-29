package types

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"math"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/blood/glucose"
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
)

type GlucosePeriods map[string]*GlucosePeriod

type GlucoseStats struct {
	Periods       GlucosePeriods `json:"periods,omitempty" bson:"periods,omitempty"`
	OffsetPeriods GlucosePeriods `json:"offsetPeriods,omitempty" bson:"offsetPeriods,omitempty"`
}

type Range struct {
	Glucose float64 `json:"glucose,omitempty" bson:"glucose,omitempty"`
	Minutes int     `json:"minutes,omitempty" bson:"minutes,omitempty"`
	Records int     `json:"records,omitempty" bson:"records,omitempty"`

	Percent  float64 `json:"percent,omitempty" bson:"percent,omitempty"`
	Variance float64 `json:"variance,omitempty" bson:"variance,omitempty"`
}

// TODO: single letter lower case pointer receiver
func (R *Range) Add(new *Range) {
	R.Variance = R.CombineVariance(new)
	R.Glucose += new.Glucose
	R.Minutes += new.Minutes
	R.Records += new.Records

	// clear percent, we don't have required values at this stage
	R.Percent = 0
}

func (R *Range) Update(record *glucoseDatum.Glucose) {
	R.Minutes += GetDuration(record)
	R.Records++
}

func (R *Range) UpdateTotal(record *glucoseDatum.Glucose) {
	normalizedValue := *glucose.NormalizeValueForUnits(record.Value, record.Units)

	// if this is bgm data, this will return 0
	duration := GetDuration(record)

	// this must occur before the regular update as the pre-increment counters are used during calc
	if duration > 0 {
		R.Variance = R.CalculateVariance(normalizedValue, float64(duration))
		R.Glucose += normalizedValue * float64(duration)
	} else {
		R.Glucose += normalizedValue
	}

	R.Update(record)
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
	GlucoseRanges      `json:",inline" bson:",inline"`
	LastRecordDuration int `json:"lastRecordDuration,omitempty" bson:"lastRecordDuration,omitempty"`
}

// TODO: define before glucose bucket
func (R *GlucoseRanges) finalizeMinutes(wallMinutes float64, days int) {
	R.Total.Percent = float64(R.Total.Minutes) / float64(days*24*60)
	// TODO: Why 0.7? What's that magic number? Add a comment explaining the conditional
	if (wallMinutes <= minutesPerDay && R.Total.Percent > 0.7) || (wallMinutes > minutesPerDay && R.Total.Minutes > minutesPerDay) {
		R.VeryLow.Percent = float64(R.VeryLow.Minutes) / wallMinutes
		R.Low.Percent = float64(R.Low.Minutes) / wallMinutes
		R.Target.Percent = float64(R.Target.Minutes) / wallMinutes
		R.High.Percent = float64(R.High.Minutes) / wallMinutes
		R.VeryHigh.Percent = float64(R.VeryHigh.Minutes) / wallMinutes
		R.ExtremeHigh.Percent = float64(R.ExtremeHigh.Minutes) / wallMinutes
		R.AnyLow.Percent = float64(R.AnyLow.Minutes) / wallMinutes
		R.AnyHigh.Percent = float64(R.AnyHigh.Minutes) / wallMinutes
	} else {
		R.VeryLow.Percent = 0
		R.Low.Percent = 0
		R.Target.Percent = 0
		R.High.Percent = 0
		R.VeryHigh.Percent = 0
		R.ExtremeHigh.Percent = 0
		R.AnyLow.Percent = 0
		R.AnyHigh.Percent = 0
	}
}

func (R *GlucoseRanges) finalizeRecords() {
	R.Total.Percent = float64(R.Total.Records) / float64(R.Total.Records)
	R.VeryLow.Percent = float64(R.VeryLow.Records) / float64(R.Total.Records)
	R.Low.Percent = float64(R.Low.Records) / float64(R.Total.Records)
	R.Target.Percent = float64(R.Target.Records) / float64(R.Total.Records)
	R.High.Percent = float64(R.High.Records) / float64(R.Total.Records)
	R.VeryHigh.Percent = float64(R.VeryHigh.Records) / float64(R.Total.Records)
	R.ExtremeHigh.Percent = float64(R.ExtremeHigh.Records) / float64(R.Total.Records)
	R.AnyLow.Percent = float64(R.AnyLow.Records) / float64(R.Total.Records)
	R.AnyHigh.Percent = float64(R.AnyHigh.Records) / float64(R.Total.Records)
}

func (R *GlucoseRanges) Finalize(state CalcState, days int) {
	if R.Total.Minutes != 0 {
		// if our bucket (period, at this point) has minutes
		wallMinutes := state.LastData.Sub(state.FirstData).Minutes() + float64(state.LastRecordDuration)
		R.finalizeMinutes(wallMinutes, days)
	} else if R.Total.Records != 0 {
		// otherwise, we only have record counts
		R.finalizeRecords()
	}
}

func (R *GlucoseRanges) Update(record *glucoseDatum.Glucose) {
	normalizedValue := *glucose.NormalizeValueForUnits(record.Value, record.Units)

	if normalizedValue < veryLowBloodGlucose {
		R.VeryLow.Update(record)
		R.AnyLow.Update(record)
	} else if normalizedValue > veryHighBloodGlucose {
		R.VeryHigh.Update(record)
		R.AnyHigh.Update(record)

		// VeryHigh is inclusive of extreme high, this is intentional
		if normalizedValue >= extremeHighBloodGlucose {
			R.ExtremeHigh.Update(record)
		}
	} else if normalizedValue < lowBloodGlucose {
		R.Low.Update(record)
		R.AnyLow.Update(record)
	} else if normalizedValue > highBloodGlucose {
		R.AnyHigh.Update(record)
		R.High.Update(record)
	} else {
		R.Target.Update(record)
	}

	R.Total.UpdateTotal(record)
}

// TODO: Glucose bucket doesn't need shared bucket.
// TODO: It needs a way to calculate blackout window. The caller should pass a blackout window calculator
func (B *GlucoseBucket) Update(r data.Datum, shared *BucketShared) (bool, error) {
	record, ok := r.(*glucoseDatum.Glucose)
	if !ok {
		return false, errors.New("record for calculation is not compatible with Glucose type")
	}

	// TODO: Update branching logic somehow? Move to a separate function
	// if we have cgm data, we care about blackout periods
	if shared.Type == SummaryTypeCGM {
		// calculate blackoutWindow based on duration of previous value
		// TODO: Magic value. Why 10 seconds?
		blackoutWindow := time.Duration(B.LastRecordDuration)*time.Minute - 10*time.Second

		// Skip record if we are within the blackout window
		if record.Time.Sub(shared.LastData) < blackoutWindow {
			return false, nil
		}
	}

	B.GlucoseRanges.Update(record)

	B.LastRecordDuration = GetDuration(record)

	return true, nil
}

type GlucosePeriod struct {
	GlucoseRanges `json:",inline" bson:",inline"`
	HoursWithData int `json:"hoursWithData,omitempty" bson:"hoursWithData,omitempty"`
	DaysWithData  int `json:"daysWithData,omitempty" bson:"daysWithData,omitempty"`

	AverageGlucose             float64 `json:"averageGlucoseMmol,omitempty" bson:"avgGlucose,omitempty"`
	GlucoseManagementIndicator float64 `json:"glucoseManagementIndicator,omitempty" bson:"GMI,omitempty"`

	CoefficientOfVariation float64 `json:"coefficientOfVariation,omitempty" bson:"CV,omitempty"`
	StandardDeviation      float64 `json:"standardDeviation,omitempty" bson:"SD,omitempty"`

	AverageDailyRecords float64 `json:"averageDailyRecords,omitempty" b;son:"avgDailyRecords,omitempty,omitempty"`

	Delta *GlucosePeriod `json:"delta,omitempty" bson:"delta,omitempty"`

	state CalcState
}

// TODO: single letter lower case pointer receiver
func (P *GlucosePeriod) Update(bucket *Bucket[*GlucoseBucket, GlucoseBucket]) error {
	if P.state.Final {
		return errors.New("period has been finalized, cannot add any data")
	}

	if bucket.Data.Total.Records == 0 {
		return nil
	}

	// TODO: check order in caller
	// NOTE this works correctly for buckets in forward or backwards order, but not unordered, it must be added with consistent direction
	// TODO: make tickets
	// NOTE this could use some math with firstData/lastData to work with non-hourly buckets, but today they're hourly.
	// NOTE should this be moved to a generic periods type as a Shared sidecar, days/hours is probably useful to other types

	if P.state.LastCountedDay.IsZero() {
		P.state.FirstCountedDay = bucket.Time
		P.state.LastCountedDay = bucket.Time

		P.state.FirstCountedHour = bucket.Time
		P.state.LastCountedHour = bucket.Time

		P.state.FirstData = bucket.FirstData
		P.state.LastData = bucket.LastData

		P.state.LastRecordDuration = bucket.Data.LastRecordDuration

		P.DaysWithData++
		P.HoursWithData++
	} else {
		if bucket.Time.Before(P.state.FirstCountedHour) {
			P.HoursWithData++
			P.state.FirstCountedHour = bucket.Time
			P.state.FirstData = bucket.FirstData

			if P.state.FirstCountedDay.Sub(bucket.Time).Hours() >= 24 {
				P.state.FirstCountedDay = bucket.Time
				P.DaysWithData++
			}
		} else if bucket.Time.After(P.state.LastCountedHour) {
			P.HoursWithData++
			P.state.LastCountedHour = bucket.Time
			P.state.LastData = bucket.LastData
			P.state.LastRecordDuration = bucket.Data.LastRecordDuration

			if bucket.Time.Sub(P.state.LastCountedDay).Hours() >= 24 {
				P.state.LastCountedDay = bucket.Time
				P.DaysWithData++
			}
		} else {
			return fmt.Errorf("bucket of time %s is within the existing period range of %s - %s",
				bucket.Time, P.state.FirstCountedHour, P.state.LastCountedHour)
		}
	}

	P.Add(&bucket.Data.GlucoseRanges)

	return nil
}

func (P *GlucosePeriod) Finalize(days int) {
	if P.state.Final != false {
		return
	}
	P.GlucoseRanges.Finalize(P.state, days)

	// if we have no records or minutes
	if P.Total.Minutes != 0 {
		P.AverageGlucose = P.Total.Glucose / float64(P.Total.Minutes)

		// we only add GMI if cgm use >70%
		if P.Total.Percent > 0.7 {
			P.GlucoseManagementIndicator = CalculateGMI(P.AverageGlucose)
		} else {
			P.GlucoseManagementIndicator = 0
		}

		P.StandardDeviation = math.Sqrt(P.Total.Variance / float64(P.Total.Minutes))
		P.CoefficientOfVariation = P.StandardDeviation / P.AverageGlucose
	} else if P.Total.Records != 0 {
		P.AverageGlucose = P.Total.Glucose / float64(P.Total.Records)
	}

	if P.Total.Records != 0 {
		P.AverageDailyRecords = float64(P.Total.Records) / float64(days)
	}

	P.state.Final = true
}

func (s *GlucoseStats) Init() {
	s.Periods = make(map[string]*GlucosePeriod)
	s.OffsetPeriods = make(map[string]*GlucosePeriod)
}

func (s *GlucoseStats) Update(ctx context.Context, bucketsCursor *mongo.Cursor) error {
	// count backwards (newest first) through hourly stats, stopping at 1d, 7d, 14d, 30d,
	// currently only supports day precision
	nextStopPoint := 0
	nextOffsetStopPoint := 0
	totalStats := GlucosePeriod{}
	totalOffsetStats := GlucosePeriod{}
	bucket := &Bucket[*GlucoseBucket, GlucoseBucket]{}

	var stopPoints []time.Time
	var offsetStopPoints []time.Time

	for bucketsCursor.Next(ctx) {
		if err := bucketsCursor.Decode(bucket); err != nil {
			return err
		}

		// Use the newest (last) bucket here to calculate date ranges
		if stopPoints == nil {
			stopPoints, offsetStopPoints = calculateStopPoints(bucket.Time)
		}

		if bucket.Data.Total.Records == 0 {
			panic("bucket exists with 0 records")
		}

		if len(stopPoints) > nextStopPoint && bucket.Time.Compare(stopPoints[nextStopPoint]) <= 0 {
			s.CalculatePeriod(periodLengths[nextStopPoint], totalStats)
			nextStopPoint++
		}

		if len(offsetStopPoints) > nextOffsetStopPoint && bucket.Time.Compare(offsetStopPoints[nextOffsetStopPoint]) <= 0 {
			s.CalculateOffsetPeriod(periodLengths[nextOffsetStopPoint], totalOffsetStats)
			nextOffsetStopPoint++
			totalOffsetStats = GlucosePeriod{}
		}

		// only count primary stats when the next stop point is a real period
		if len(stopPoints) > nextStopPoint {
			if err := totalStats.Update(bucket); err != nil {
				return err
			}
		}

		// only add to offset stats when primary stop point is ahead of offset
		if nextStopPoint > nextOffsetStopPoint && len(offsetStopPoints) > nextOffsetStopPoint {
			if err := totalOffsetStats.Update(bucket); err != nil {
				return err
			}
		}
	}

	// fill in periods we never reached
	for i := nextStopPoint; i < len(stopPoints); i++ {
		s.CalculatePeriod(periodLengths[i], totalStats)
	}
	for i := nextOffsetStopPoint; i < len(offsetStopPoints); i++ {
		s.CalculateOffsetPeriod(periodLengths[i], totalOffsetStats)
		totalOffsetStats = GlucosePeriod{}
	}

	s.CalculateDelta()

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
		Delta(&s.Periods[k].GlucoseManagementIndicator, &s.OffsetPeriods[k].GlucoseManagementIndicator, &s.Periods[k].Delta.GlucoseManagementIndicator, &s.OffsetPeriods[k].Delta.GlucoseManagementIndicator)
		Delta(&s.Periods[k].AverageDailyRecords, &s.OffsetPeriods[k].AverageDailyRecords, &s.Periods[k].Delta.AverageDailyRecords, &s.OffsetPeriods[k].Delta.AverageDailyRecords)
		Delta(&s.Periods[k].StandardDeviation, &s.OffsetPeriods[k].StandardDeviation, &s.Periods[k].Delta.StandardDeviation, &s.OffsetPeriods[k].Delta.StandardDeviation)
		Delta(&s.Periods[k].CoefficientOfVariation, &s.OffsetPeriods[k].CoefficientOfVariation, &s.Periods[k].Delta.CoefficientOfVariation, &s.OffsetPeriods[k].Delta.CoefficientOfVariation)
		Delta(&s.Periods[k].DaysWithData, &s.OffsetPeriods[k].DaysWithData, &s.Periods[k].Delta.DaysWithData, &s.OffsetPeriods[k].Delta.DaysWithData)
		Delta(&s.Periods[k].HoursWithData, &s.OffsetPeriods[k].HoursWithData, &s.Periods[k].Delta.HoursWithData, &s.OffsetPeriods[k].Delta.HoursWithData)
	}
}

func (s *GlucoseStats) CalculatePeriod(days int, period GlucosePeriod) {
	period.Finalize(days)
	s.Periods[strconv.Itoa(days)+"d"] = &period
}

func (s *GlucoseStats) CalculateOffsetPeriod(days int, period GlucosePeriod) {
	period.Finalize(days)
	s.OffsetPeriods[strconv.Itoa(days)+"d"] = &period
}
