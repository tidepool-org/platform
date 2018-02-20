package test

import (
	"github.com/tidepool-org/platform/data"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewSuppressedScheduled() *basal.Suppressed {
	datum := basal.NewSuppressed()
	datum.Type = pointer.String("basal")
	datum.DeliveryType = pointer.String("scheduled")
	datum.Annotations = testData.NewBlobArray()
	datum.Rate = pointer.Float64(test.RandomFloat64FromRange(basal.RateMinimum, basal.RateMaximum))
	datum.ScheduleName = pointer.String(NewScheduleName())
	return datum
}

func NewSuppressedSuspend() *basal.Suppressed { // Not legal, but used for negative unit tests
	datum := basal.NewSuppressed()
	datum.Type = pointer.String("basal")
	datum.DeliveryType = pointer.String("suspend")
	datum.Annotations = testData.NewBlobArray()
	datum.Rate = pointer.Float64(test.RandomFloat64FromRange(basal.RateMinimum, basal.RateMaximum))
	return datum
}

func NewSuppressedTemporary() *basal.Suppressed {
	datum := basal.NewSuppressed()
	datum.Type = pointer.String("basal")
	datum.DeliveryType = pointer.String("temp")
	datum.Annotations = testData.NewBlobArray()
	datum.Rate = pointer.Float64(test.RandomFloat64FromRange(basal.RateMinimum, basal.RateMaximum))
	datum.Suppressed = basal.NewSuppressed()
	datum.Suppressed.Type = pointer.String("basal")
	datum.Suppressed.DeliveryType = pointer.String("scheduled")
	datum.Suppressed.Annotations = testData.NewBlobArray()
	datum.Suppressed.Rate = pointer.Float64(test.RandomFloat64FromRange(basal.RateMinimum, basal.RateMaximum))
	datum.Suppressed.ScheduleName = pointer.String(NewScheduleName())
	return datum
}

func CloneSuppressed(datum *basal.Suppressed) *basal.Suppressed {
	if datum == nil {
		return nil
	}
	clone := basal.NewSuppressed()
	clone.Type = test.CloneString(datum.Type)
	clone.DeliveryType = test.CloneString(datum.DeliveryType)
	clone.Annotations = testData.CloneBlobArray(datum.Annotations)
	clone.Rate = test.CloneFloat64(datum.Rate)
	clone.ScheduleName = test.CloneString(datum.ScheduleName)
	clone.Suppressed = CloneSuppressed(datum.Suppressed)
	return clone
}

func NewTestSuppressed(typ interface{}, deliveryType interface{}, annotations *data.BlobArray, rate interface{}, scheduleName interface{}, suppressed *basal.Suppressed) *basal.Suppressed { // TODO: Remove once Parse tests are updated
	datum := &basal.Suppressed{}
	if val, ok := typ.(string); ok {
		datum.Type = &val
	}
	if val, ok := deliveryType.(string); ok {
		datum.DeliveryType = &val
	}
	datum.Annotations = annotations
	if val, ok := rate.(float64); ok {
		datum.Rate = &val
	}
	if val, ok := scheduleName.(string); ok {
		datum.ScheduleName = &val
	}
	datum.Suppressed = suppressed
	return datum
}
