package api

import "github.com/tidepool-org/platform/dataservices/service"

type Status struct {
	Version     string
	Environment string
	DataStore   interface{}
	Server      interface{}
}

func (s *Standard) GetStatus(serviceContext service.Context) {
	status := &Status{
		Version:     s.VersionReporter().Long(),
		Environment: s.EnvironmentReporter().Name(),
		DataStore:   s.dataStore.GetStatus(),
		Server:      s.StatusMiddleware().GetStatus(),
	}
	serviceContext.Response().WriteJson(status)
}
