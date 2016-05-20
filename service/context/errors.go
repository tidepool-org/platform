package context

import (
	"net/http"

	"github.com/tidepool-org/platform/service"
)

func ErrorInternalServerFailure() *service.Error {
	return &service.Error{
		Code:   "internal-server-failure",
		Status: http.StatusInternalServerError,
		Title:  "internal server failure",
		Detail: "Internal server failure",
	}
}
