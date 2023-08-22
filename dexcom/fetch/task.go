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

const dexcomTaskRetryField = "retryCount"

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
			"providerSessionId":  providerSessionID,
			"dataSourceId":       dataSourceID,
			dexcomTaskRetryField: 0,
		},
	}, nil
}

func SetErrorOrAllowTaskRetry(t *task.Task, err error) {
	if shouldTaskError(t) {
		t.AppendError(err)
		t.SetFailed()
		return
	}
	incrementTaskRetryCount(t)
}

func shouldTaskError(t *task.Task) bool {
	if t.Data[dexcomTaskRetryField] != nil {
		count, ok := t.Data[dexcomTaskRetryField].(int)
		if ok {
			return count > 3
		}
	}
	return true
}

func incrementTaskRetryCount(t *task.Task) {
	if t.Data[dexcomTaskRetryField] != nil {
		count, ok := t.Data[dexcomTaskRetryField].(int)
		if ok {
			t.Data[dexcomTaskRetryField] = count + 1
		}
	} else {
		t.Data[dexcomTaskRetryField] = 1
	}
}
