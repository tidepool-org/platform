package api

import "github.com/tidepool-org/platform/userservices/service"

type Version struct {
	Version string `json:"version"`
}

func (s *Standard) GetVersion(serviceContext service.Context) {
	serviceContext.Response().WriteJson(Version{s.VersionReporter().Long()})
}
