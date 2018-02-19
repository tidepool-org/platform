package test

import (
	"time"

	"github.com/tidepool-org/platform/data/types/bolus/extended"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewExtended(typ string, subType string, duration interface{}, durationExpected interface{}, extndd interface{}, extnddExpected interface{}) *extended.Extended {
	datum := extended.Init()
	datum.CreatedTime = pointer.String(time.Now().Format(time.RFC3339))
	datum.DeviceID = pointer.String(id.New())
	datum.Time = pointer.String(test.NewTime().Format(time.RFC3339))
	datum.UploadID = pointer.String(id.New())
	datum.UserID = pointer.String(id.New())
	datum.Type = typ
	datum.SubType = subType
	if val, ok := duration.(int); ok {
		datum.Duration = &val
	}
	if val, ok := durationExpected.(int); ok {
		datum.DurationExpected = &val
	}
	if val, ok := extndd.(float64); ok {
		datum.Extended = &val
	}
	if val, ok := extnddExpected.(float64); ok {
		datum.ExtendedExpected = &val
	}
	return datum
}
