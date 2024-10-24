package task

import (
	"time"

	"github.com/tidepool-org/platform/pointer"
)

const CarePartnerType = "org.tidepool.carepartner"

func NewCarePartnerTaskCreate() *TaskCreate {
	return &TaskCreate{
		Name:          pointer.FromAny(CarePartnerType),
		Type:          CarePartnerType,
		AvailableTime: &time.Time{},
		Data:          map[string]interface{}{},
	}
}
