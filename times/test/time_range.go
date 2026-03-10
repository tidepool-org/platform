package test

import (
	"time"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/times"
)

func RandomTimeRange(options ...test.Option) *times.TimeRange {
	datum := &times.TimeRange{}
	datum.From = test.RandomOptional(test.RandomTime, options...)
	datum.To = test.RandomOptional(func() time.Time { return test.RandomTimeAfter(pointer.Default(datum.From, time.Time{})) }, options...)
	return datum
}

func CloneTimeRange(datum *times.TimeRange) *times.TimeRange {
	if datum == nil {
		return nil
	}
	return &times.TimeRange{
		From: pointer.Clone(datum.From),
		To:   pointer.Clone(datum.To),
	}
}

func NewObjectFromTimeRange(datum *times.TimeRange, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.From != nil {
		object["from"] = test.NewObjectFromTime(*datum.From, objectFormat)
	}
	if datum.To != nil {
		object["to"] = test.NewObjectFromTime(*datum.To, objectFormat)
	}
	return object
}

func RandomTimeRangeMetadata(options ...test.Option) *times.TimeRangeMetadata {
	return &times.TimeRangeMetadata{
		TimeRange: RandomTimeRange(options...),
	}
}

func CloneTimeRangeMetadata(datum *times.TimeRangeMetadata) *times.TimeRangeMetadata {
	if datum == nil {
		return nil
	}
	return &times.TimeRangeMetadata{
		TimeRange: CloneTimeRange(datum.TimeRange),
	}
}

func NewObjectFromTimeRangeMetadata(datum *times.TimeRangeMetadata, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.TimeRange != nil {
		object["timeRange"] = NewObjectFromTimeRange(datum.TimeRange, objectFormat)
	}
	return object
}
