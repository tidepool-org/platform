package summary

import (
	"fmt"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
)

func TaskName() string {
	return fmt.Sprintf("%s", Type)
}

func NewTaskCreate() (*task.TaskCreate, error) {
	return &task.TaskCreate{
		Name: pointer.FromString(TaskName()),
		Type: Type,
		Data: map[string]interface{}{},
	}, nil
}
