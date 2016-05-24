package context

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"fmt"
	"strings"

	"github.com/tidepool-org/platform/data"
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
