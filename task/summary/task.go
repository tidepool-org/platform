package summary

import (
	"fmt"
	"time"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
)

func TaskName() string {
	return fmt.Sprintf("%s", Type)
}

func NewDefaultTaskCreate() *task.TaskCreate {
	availableTime := time.Now().UTC()
	expirationTime := availableTime.AddDate(1000, 0, 0)

	return &task.TaskCreate{
		Name:           pointer.FromString(TaskName()),
		Type:           Type,
		Priority:       5,
		AvailableTime:  pointer.FromTime(availableTime),
		ExpirationTime: pointer.FromTime(expirationTime),
	}
}
