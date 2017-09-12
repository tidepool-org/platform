package api

import (
	dataservicesService "github.com/tidepool-org/platform/dataservices/service"
	"github.com/tidepool-org/platform/service"
)

type Status struct {
	Version     string
	Environment string
	DataStore   interface{}
	Server      interface{}
}

func (s *Standard) GetStatus(serviceContext dataservicesService.Context) {
	status := &Status{
		Version:     s.VersionReporter().Long(),
		Environment: s.EnvironmentReporter().Name(),
		DataStore:   s.dataStore.GetStatus(),
		Server:      s.StatusMiddleware().GetStatus(),
	}
	service.AddDateHeader(serviceContext.Response())
	serviceContext.Response().WriteJson(status)
}
