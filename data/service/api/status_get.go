package api

import (
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
)

type Status struct {
	Version     string
	Environment string
	DataStore   interface{}
	Server      interface{}
}

func (s *Standard) StatusGet(dataServiceContext dataService.Context) {
	status := &Status{
		Version: s.VersionReporter().Long(),
	}
	dataServiceContext.RespondWithStatusAndData(http.StatusOK, status)
}
