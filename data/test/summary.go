package test

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/test"
)

func RandomSummary() *data.Summary {
	datum := data.Summary{}
	datum.LastUpdated = test.RandomTime()
	datum.LastUpload = test.RandomTime()
    datum.LastData = test.RandomTime()
	datum.AverageGlucose = test.RandomFloat64FromRange(1, 15)
	datum.TimeInRange = test.RandomFloat64FromRange(0, 1)

	return &datum
}
