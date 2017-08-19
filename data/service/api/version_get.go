package api

import (
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
)

type Version struct {
	Version string `json:"version"`
}

func (s *Standard) VersionGet(dataServiceContext dataService.Context) {
	dataServiceContext.RespondWithStatusAndData(http.StatusOK, Version{s.VersionReporter().Long()})
}
