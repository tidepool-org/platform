package sync

import (
	"fmt"
	"time"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
)

const (
	Type = "org.tidepool.ehr.sync"
)

func TaskName(clinicId string) string {
	return fmt.Sprintf("%s:%s", Type, clinicId)
}

func NewTaskCreate(clinicId string) *task.TaskCreate {
	return &task.TaskCreate{
		Name:          pointer.FromString(TaskName(clinicId)),
		Type:          Type,
		AvailableTime: pointer.FromAny(time.Now().UTC()),
		Data: map[string]interface{}{
			"clinicId": clinicId,
		},
	}
}

func GetClinicId(data map[string]interface{}) (string, error) {
	clinicId, ok := data["clinicId"].(string)
	if !ok {
		return "", fmt.Errorf("unable to get clinicId from task data")
	}
	return clinicId, nil
}
