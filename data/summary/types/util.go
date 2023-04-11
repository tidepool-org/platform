package types

import (
	"math"
	"strings"
	"time"

	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
)

func SkipUntil[T RecordTypes, A RecordTypesPt[T]](date time.Time, userData []A) ([]A, error) {
	var skip int
	for i := 0; i < len(userData); i++ {
		recordTime := userData[i].GetTime()

		if recordTime.Before(date) {
			skip = i + 1
		} else {
			break
		}
	}

	if skip > 0 {
		userData = userData[skip:]
	}

	return userData, nil
}

// GetDuration assumes all except freestyle is 5 minutes
func GetDuration(dataSet *glucoseDatum.Glucose) int {
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

// CalculateRealMinutes remove partial hour (data end) from total time for more accurate TimeCGMUse
func CalculateRealMinutes(i int, lastRecordTime time.Time) float64 {
	realMinutes := float64(i * 24 * 60)
	nextHour := time.Date(lastRecordTime.Year(), lastRecordTime.Month(), lastRecordTime.Day(),
		lastRecordTime.Hour()+1, 0, 0, 0, lastRecordTime.Location())
	realMinutes = realMinutes - nextHour.Sub(lastRecordTime).Minutes()

	return realMinutes
}
