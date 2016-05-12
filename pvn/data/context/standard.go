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
	if reference == nil || err == nil {
		return
	}

	s.Errors.AppendError(err.WithPointer(s.appendReference(reference)))
}

func (s *Standard) NewChildContext(reference interface{}) data.Context {
	if reference == nil {
		return nil
	}

	return &Standard{
		Errors:  s.Errors,
		pointer: s.appendReference(reference),
	}
}

func (s *Standard) appendReference(reference interface{}) string {
	pointer := fmt.Sprintf("%v", reference)
	if s.pointer != "" {
		pointer = strings.Join([]string{s.pointer, pointer}, "/")
	}
	return pointer
}
