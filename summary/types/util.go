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
func GetDuration(datum *glucoseDatum.Glucose) int {
	if datum.Type != continuous.Type {
		// non-continuous has no duration
		return 0
	}
	if datum.DeviceID != nil {
		if strings.Contains(*datum.DeviceID, "AbbottFreeStyleLibre3") {
			return 5
		}
		if strings.Contains(*datum.DeviceID, "AbbottFreeStyleLibre") {
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

func BinDelta(currentRange, offsetRange, deltaRange *Range) {
	deltaRange.Percent = currentRange.Percent - offsetRange.Percent
	deltaRange.Records = currentRange.Records - offsetRange.Records
	deltaRange.Minutes = currentRange.Minutes - offsetRange.Minutes
}

func Delta[T Number](current, previous, destination *T) {
	*destination = *current - *previous
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
