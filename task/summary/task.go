package summary

import (
	"fmt"
	"time"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
)

func BackfillTaskName() string {
	return fmt.Sprintf("%s", BackfillType)
}

func UpdateTaskName() string {
	return fmt.Sprintf("%s", UpdateType)
}

func NewDefaultBackfillTaskCreate() *task.TaskCreate {
	availableTime := time.Now().UTC()
	expirationTime := availableTime.AddDate(1000, 0, 0)

	return &task.TaskCreate{
		Name:           pointer.FromString(BackfillTaskName()),
		Type:           BackfillType,
		Priority:       5,
		AvailableTime:  pointer.FromTime(availableTime),
		ExpirationTime: pointer.FromTime(expirationTime),
	}
}

func NewDefaultUpdateTaskCreate() *task.TaskCreate {
	availableTime := time.Now().UTC()
	expirationTime := availableTime.AddDate(1000, 0, 0)

	return &task.TaskCreate{
		Name:           pointer.FromString(UpdateTaskName()),
		Type:           UpdateType,
		Priority:       5,
		AvailableTime:  pointer.FromTime(availableTime),
		ExpirationTime: pointer.FromTime(expirationTime),
	}
}
