package v1

import (
	"net/http"
	"time"

	"github.com/tidepool-org/platform/data/service"
)

type timeGet struct {
	Time string `json:"time"`
}

// TimeGet godoc
// @Summary Get current server time
// @ID platform-data-api-TimeGet
// @Produce json
// @Success 200 {object} timeGet "Current time with format RFC3339Nano (2006-01-02T15:04:05.999999999Z07:00)"
// @Router /v1/time [get]
func TimeGet(serviceContext service.Context) {
	response := timeGet{
		Time: time.Now().Truncate(time.Millisecond).Format(time.RFC3339Nano),
	}
	serviceContext.RespondWithStatusAndData(http.StatusOK, response)
}
