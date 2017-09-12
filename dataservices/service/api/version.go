package api

import (
	dataservicesService "github.com/tidepool-org/platform/dataservices/service"
	"github.com/tidepool-org/platform/service"
)

type Version struct {
	Version string `json:"version"`
}

func (s *Standard) GetVersion(serviceContext dataservicesService.Context) {
	service.AddDateHeader(serviceContext.Response())
	serviceContext.Response().WriteJson(Version{s.VersionReporter().Long()})
}
