package api

import userService "github.com/tidepool-org/platform/user/service"

type Version struct {
	Version string `json:"version"`
}

func (s *Standard) GetVersion(userServiceContext userService.Context) {
	userServiceContext.Response().WriteJson(Version{s.VersionReporter().Long()})
}
