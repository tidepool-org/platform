package reconcile

import (
	"time"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
)

const (
	Type = "org.tidepool.ehr.reconcile"
)

func NewTaskCreate() *task.TaskCreate {
	return &task.TaskCreate{
		Name:          pointer.FromString(Type),
		Type:          Type,
		AvailableTime: pointer.FromAny(time.Now().UTC()),
	}
}
