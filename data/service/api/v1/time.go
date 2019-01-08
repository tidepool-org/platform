package v1

import (
	"net/http"
	"time"

	"github.com/tidepool-org/platform/data/service"
)

func TimeGet(serviceContext service.Context) {
	response := struct {
		Time string `json:"time"`
	}{
		Time: time.Now().Truncate(time.Millisecond).Format(time.RFC3339Nano),
	}
	serviceContext.RespondWithStatusAndData(http.StatusOK, response)
}
