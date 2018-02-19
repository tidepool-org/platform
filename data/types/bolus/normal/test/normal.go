package test

import (
	"time"

	"github.com/tidepool-org/platform/data/types/bolus/normal"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewNormal(typ string, subType string, nrml interface{}, nrmlExpected interface{}) *normal.Normal {
	datum := normal.Init()
	datum.CreatedTime = pointer.String(time.Now().Format(time.RFC3339))
	datum.DeviceID = pointer.String(id.New())
	datum.Time = pointer.String(test.NewTime().Format(time.RFC3339))
	datum.UploadID = pointer.String(id.New())
	datum.UserID = pointer.String(id.New())
	datum.Type = typ
	datum.SubType = subType
	if val, ok := nrml.(float64); ok {
		datum.Normal = &val
	}
	if val, ok := nrmlExpected.(float64); ok {
		datum.NormalExpected = &val
	}
	return datum
}
