package api

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import "github.com/tidepool-org/platform/dataservices/service"

type Status struct {
	Version     string
	Environment string
	Store       interface{}
	Server      interface{}
}

func (s *Standard) GetStatus(serviceContext service.Context) {
	status := &Status{
		Version:     s.versionReporter.Long(),
		Environment: s.environmentReporter.Name(),
		Store:       s.dataStore.GetStatus(),
		Server:      s.statusMiddleware.GetStatus(),
	}
	serviceContext.Response().WriteJson(status)
}
