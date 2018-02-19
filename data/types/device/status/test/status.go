package test

import (
	"time"

	"github.com/tidepool-org/platform/data/types/device/status"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewReason() *map[string]interface{} {
	return testDataTypes.NewPropertyMap()
}

func NewStatus(typ string, subType string, duration interface{}, name interface{}, reason *map[string]interface{}) *status.Status {
	datum := status.Init()
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
	if val, ok := name.(string); ok {
		datum.Name = &val
	}
	datum.Reason = reason
	return datum
}
