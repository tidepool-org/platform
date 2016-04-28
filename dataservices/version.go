package dataservices

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/version"
)

type Version struct {
	Version string `json:"version"`
}

func (s *Server) GetVersion(response rest.ResponseWriter, request *rest.Request) {
	response.WriteJson(Version{version.Long()})
}
