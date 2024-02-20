package fetch

import (
	"fmt"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
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

// TODO: DARIN - Not sure about this logic
func ErrorOrRetryTask(t *task.Task, err error) {
	t.AppendError(err)
	if t.IsFailed() {
		if shouldTaskFail(t) {
			return
		}
		incrementTaskRetryCount(t)
		t.State = task.TaskStateRunning // TODO: DARIN - calls RepeatAvailableAfter later which sets Pending
	}
}

func FailTask(l log.Logger, t *task.Task, err error) error {
	l.WithError(err).Warnf("dexcom task %s failed", t.ID)
	t.SetFailed()
	return err
}

func shouldTaskFail(t *task.Task) bool {
	if t.Data[dexcomTaskRetryField] != nil {
		count, ok := t.Data[dexcomTaskRetryField].(int)
		if ok {
			return count >= 3
		}
	}
	return true
}

func incrementTaskRetryCount(t *task.Task) {
	if t.Data[dexcomTaskRetryField] != nil {
		count, ok := t.Data[dexcomTaskRetryField].(int)
		if ok {
			count++
			t.Data[dexcomTaskRetryField] = count
		}
		// TODO: What if wrong? Should reset?
	} else {
		t.Data[dexcomTaskRetryField] = 1
	}
}
