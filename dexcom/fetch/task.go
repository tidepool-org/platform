package fetch

import (
	"fmt"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
)

func TaskName(providerSessionID string) string {
	return fmt.Sprintf("%s:%s", Type, providerSessionID)
}

func NewTaskCreate(providerSessionID string, dataSourceID string) (*task.TaskCreate, error) {
	if providerSessionID == "" {
		return nil, errors.New("provider session id is missing")
	}
	if dataSourceID == "" {
		return nil, errors.New("data source id is missing")
	}

	return &task.TaskCreate{
		Name: pointer.FromString(TaskName(providerSessionID)),
		Type: Type,
		Data: map[string]interface{}{
			"providerSessionId": providerSessionID,
			"dataSourceId":      dataSourceID,
		},
	}, nil
}
