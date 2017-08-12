package api

import (
	"net/http"

	userService "github.com/tidepool-org/platform/user/service"
)

type Version struct {
	Version string `json:"version"`
}

func (s *Standard) GetVersion(userServiceContext userService.Context) {
	userServiceContext.RespondWithStatusAndData(http.StatusOK, Version{s.VersionReporter().Long()})
}
