package types

import (
	"math"
	"strings"
	"time"

	"golang.org/x/exp/constraints"

	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
)

// GetDuration assumes all except freestyle is 5 minutes
func GetDuration(dataSet *glucoseDatum.Glucose) int {
	if dataSet.Type != continuous.Type {
		// non-continuous has no duration
		return 0
	}
	if dataSet.DeviceID != nil {
		if strings.Contains(*dataSet.DeviceID, "AbbottFreeStyleLibre3") {
			return 5
		}
		if strings.Contains(*dataSet.DeviceID, "AbbottFreeStyleLibre") {
			return 15
		}
	}
	return 5
}

func CalculateGMI(averageGlucose float64) float64 {
	gmi := 12.71 + 4.70587*averageGlucose
	gmi = (0.09148 * gmi) + 2.152
	gmi = math.Round(gmi*10) / 10
	return gmi
}

// CalculateWallMinutes remove partial hour (data end) from total time for more accurate percentages
func CalculateWallMinutes(i int, lastRecordTime time.Time, lastRecordDuration int) float64 {
	realMinutes := float64(i * 24 * 60)
	nextHour := time.Date(lastRecordTime.Year(), lastRecordTime.Month(), lastRecordTime.Day(),
		lastRecordTime.Hour()+1, 0, 0, 0, lastRecordTime.Location())
	potentialRealMinutes := realMinutes - nextHour.Sub(lastRecordTime.Add(time.Minute*time.Duration(lastRecordDuration))).Minutes()

	if potentialRealMinutes > realMinutes {
		return realMinutes
	}
	return potentialRealMinutes
}

type Number interface {
	constraints.Float | constraints.Integer
}

func Abs[T Number](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

func BinDelta(bin, offsetBin, deltaBin, offsetDeltaBin *Range) {
	deltaBin.Percent = bin.Percent - offsetBin.Percent
	offsetDeltaBin.Percent = -deltaBin.Percent

	deltaBin.Records = bin.Records - offsetBin.Records
	offsetDeltaBin.Records = -deltaBin.Records

	deltaBin.Minutes = bin.Minutes - offsetBin.Minutes
	offsetDeltaBin.Minutes = -deltaBin.Minutes
}

func Delta[T Number](a, b, c, d *T) {
	*c = *a - *b
	*d = -*c
}

var periodLengths = [...]int{1, 7, 14, 30}

func calculateStopPoints(endTime time.Time) ([]time.Time, []time.Time) {
	stopPoints := make([]time.Time, len(periodLengths))
	offsetStopPoints := make([]time.Time, len(periodLengths))
	for i := range periodLengths {
		stopPoints[i] = endTime.AddDate(0, 0, -periodLengths[i])
		offsetStopPoints[i] = endTime.AddDate(0, 0, -periodLengths[i]*2)
	}

	return stopPoints, offsetStopPoints
}
