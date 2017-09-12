package api

import (
	"github.com/tidepool-org/platform/service"
	userservicesService "github.com/tidepool-org/platform/userservices/service"
)

type Version struct {
	Version string `json:"version"`
}

func (s *Standard) GetVersion(serviceContext userservicesService.Context) {
	service.AddDateHeader(serviceContext.Response())
	serviceContext.Response().WriteJson(Version{s.VersionReporter().Long()})
}
