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

type Version struct {
	Version string `json:"version"`
}

func (s *Standard) GetVersion(serviceContext service.Context) {
	serviceContext.Response().WriteJson(Version{s.VersionReporter().Long()})
}
