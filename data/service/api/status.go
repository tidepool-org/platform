package api

import dataService "github.com/tidepool-org/platform/data/service"

type Status struct {
	Version     string
	Environment string
	DataStore   interface{}
	Server      interface{}
}

func (s *Standard) GetStatus(dataServiceContext dataService.Context) {
	status := &Status{
		Version:   s.VersionReporter().Long(),
		DataStore: s.dataStore.GetStatus(),
		Server:    s.StatusMiddleware().GetStatus(),
	}
	dataServiceContext.Response().WriteJson(status)
}
