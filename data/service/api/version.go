package api

import dataService "github.com/tidepool-org/platform/data/service"

type Version struct {
	Version string `json:"version"`
}

func (s *Standard) GetVersion(dataServiceContext dataService.Context) {
	dataServiceContext.Response().WriteJson(Version{s.VersionReporter().Long()})
}
