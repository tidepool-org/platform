package summary

import (
	"math"
	"strings"
	"time"

	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
)

// GetDuration assumes all except freestyle is 5 minutes
func GetDuration(dataSet *glucoseDatum.Glucose) int {
	if dataSet.DeviceID != nil {
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

func SkipUntil(date time.Time, userData []*glucoseDatum.Glucose) ([]*glucoseDatum.Glucose, error) {
	var skip int
	for i := 0; i < len(userData); i++ {
		recordTime := *userData[i].Time

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
