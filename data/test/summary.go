package test

import (
	"github.com/tidepool-org/platform/data/types/blood/glucose/summary"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomSummary() *summary.Summary {
	datum := summary.Summary{}
	datum.LastUpdated = pointer.FromTime(test.RandomTime())
	datum.LastUpload = pointer.FromTime(test.RandomTime())
	datum.FirstData = pointer.FromTime(test.RandomTime())
	datum.LastData = pointer.FromTime(test.RandomTime())
	datum.TimeInRange = pointer.FromFloat64(test.RandomFloat64FromRange(0, 1))
	datum.TimeBelowRange = pointer.FromFloat64(test.RandomFloat64FromRange(0, 1))
	datum.TimeAboveRange = pointer.FromFloat64(test.RandomFloat64FromRange(0, 1))
	datum.LowGlucoseThreshold = pointer.FromFloat64(test.RandomFloat64FromRange(0, 5))
	datum.HighGlucoseThreshold = pointer.FromFloat64(test.RandomFloat64FromRange(5, 20))

	datum.AverageGlucose = &summary.Glucose{
		Value: pointer.FromFloat64(test.RandomFloat64FromRange(1, 30)),
		Units: pointer.FromString("mmol/l"),
	}

	return &datum
}
