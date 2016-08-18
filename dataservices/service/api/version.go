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

import "github.com/ant0ine/go-json-rest/rest"

type Version struct {
	Version string `json:"version"`
}

func (s *Standard) GetVersion(response rest.ResponseWriter, request *rest.Request) {
	response.WriteJson(Version{s.versionReporter.Long()})
}
