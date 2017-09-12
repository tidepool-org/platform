package v1

import (
	"net/http"
	"time"

	"github.com/tidepool-org/platform/dataservices/service"
)

func TimeGet(serviceContext service.Context) {
	response := struct {
		Time string `json:"time"`
	}{
		Time: time.Now().UTC().Format(time.RFC3339),
	}
	serviceContext.RespondWithStatusAndData(http.StatusOK, response)
}
