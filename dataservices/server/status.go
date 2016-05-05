package server

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/version"
)

type Status struct {
	Version string
	Store   interface{}
	Server  interface{}
}

func (s *Server) GetStatus(response rest.ResponseWriter, request *rest.Request) {
	status := &Status{
		Version: version.Current().Long(),
		Store:   s.store.GetStatus(),
		Server:  s.statusMiddleware.GetStatus(),
	}
	response.WriteJson(status)
}
