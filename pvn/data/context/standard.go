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
	errors  *[]*service.Error
	pointer string
}

func NewStandard() *Standard {
	return &Standard{
		errors: &[]*service.Error{},
	}
}

func (s *Standard) Errors() []*service.Error {
	return *s.errors
}

func (s *Standard) AppendError(reference interface{}, err *service.Error) {
	if err != nil {
		*s.errors = append(*s.errors, err.WithPointer(s.appendReference(reference)))
	}
}

func (s *Standard) NewChildContext(reference interface{}) data.Context {
	return &Standard{
		errors:  s.errors,
		pointer: s.appendReference(reference),
	}
}

func (s *Standard) appendReference(reference interface{}) string {
	return strings.Join([]string{s.pointer, fmt.Sprintf("%v", reference)}, "/")
}
