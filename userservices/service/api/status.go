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

import "github.com/tidepool-org/platform/userservices/service"

type Status struct {
	Version     string
	Environment string
	Server      interface{}
}

func (s *Standard) GetStatus(serviceContext service.Context) {
	status := &Status{
		Version:     s.VersionReporter().Long(),
		Environment: s.EnvironmentReporter().Name(),
		Server:      s.StatusMiddleware().GetStatus(),
	}
	serviceContext.Response().WriteJson(status)
}
