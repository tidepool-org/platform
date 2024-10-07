package sync

import (
	"fmt"
	"time"

	"github.com/tidepool-org/platform/errors"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
)

const (
	Type           = "org.tidepool.ehr.sync"
	DefaultCadence = time.Hour * 24 * 14
)

func TaskName(clinicId string) string {
	return fmt.Sprintf("%s:%s", Type, clinicId)
}

func NewTaskCreate(clinicId string, cadence time.Duration) *task.TaskCreate {
	tsk := &task.TaskCreate{
		Name:          pointer.FromString(TaskName(clinicId)),
		Type:          Type,
		AvailableTime: pointer.FromAny(time.Now().UTC()),
		Data: map[string]interface{}{
			"clinicId": clinicId,
		},
	}
	SetCadence(tsk.Data, cadence)
	return tsk
}

func GetClinicId(data map[string]interface{}) (string, error) {
	clinicId, ok := data["clinicId"].(string)
	if !ok {
		return "", errors.New("unable to get clinicId from task data")
	}
	return clinicId, nil
}

func ScheduleNextExecution(tsk *task.Task) {
	if tsk.HasError() {
		tsk.RepeatAvailableAfter(OnErrorAvailableAfterDuration)
		return
	}

	cadence := GetCadence(tsk.Data)
	if cadence == nil {
		tsk.RepeatAvailableAfter(DefaultCadence)
	} else if *cadence != 0 {
		tsk.RepeatAvailableAfter(*cadence)
	} else {
		tsk.SetCompleted()
	}
}

func GetCadence(data map[string]interface{}) *time.Duration {
	cadence := data["cadence"].(string)
	parsed, err := time.ParseDuration(cadence)
	if err != nil {
		return nil
	}
	return &parsed
}

func SetCadence(data map[string]interface{}, period time.Duration) {
	data["cadence"] = period.String()
}
