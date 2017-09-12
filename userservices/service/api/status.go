package api

import (
	"github.com/tidepool-org/platform/service"
	userservicesService "github.com/tidepool-org/platform/userservices/service"
)

type Status struct {
	Version     string
	Environment string
	Server      interface{}
}

func (s *Standard) GetStatus(serviceContext userservicesService.Context) {
	status := &Status{
		Version:     s.VersionReporter().Long(),
		Environment: s.EnvironmentReporter().Name(),
		Server:      s.StatusMiddleware().GetStatus(),
	}
	service.AddDateHeader(serviceContext.Response())
	serviceContext.Response().WriteJson(status)
}
