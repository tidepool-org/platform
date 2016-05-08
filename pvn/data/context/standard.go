package context

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
	"fmt"
	"strings"

	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/service"
)

type Standard struct {
	*service.Errors
	pointer string
}

func NewStandard() *Standard {
	return &Standard{
		Errors: service.NewErrors(),
	}
}

func (s *Standard) AppendError(reference interface{}, err *service.Error) {
	s.Errors.AppendError(err.WithPointer(joinStringReferences(s.pointer, stringifyReference(reference))))
}

func (s *Standard) NewChildContext(reference interface{}) data.Context {
	pointer := stringifyReference(reference)
	if s.pointer != "" {
		pointer = joinStringReferences(s.pointer, pointer)
	}
	return &Standard{
		Errors:  s.Errors,
		pointer: pointer,
	}
}

func stringifyReference(reference interface{}) string {
	return fmt.Sprintf("%v", reference)
}

func joinStringReferences(references ...string) string {
	return strings.Join(references, "/")
}
