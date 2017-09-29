package context

import (
	"fmt"
	"strings"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

type Standard struct {
	logger  log.Logger
	pointer string
	meta    interface{}
	errors  *[]*service.Error
}

func NewStandard(logger log.Logger) (*Standard, error) {
	if logger == nil {
		return nil, errors.New("logger is missing")
	}

	return &Standard{
		logger: logger,
		errors: &[]*service.Error{},
	}, nil
}

func (s *Standard) Logger() log.Logger {
	return s.logger
}

func (s *Standard) Meta() interface{} {
	return s.meta
}

func (s *Standard) SetMeta(meta interface{}) {
	s.meta = meta
}

func (s *Standard) ResolveReference(reference interface{}) string {
	return strings.Join([]string{s.pointer, fmt.Sprintf("%v", reference)}, "/")
}

func (s *Standard) Errors() []*service.Error {
	return *s.errors
}

func (s *Standard) AppendError(reference interface{}, err *service.Error) {
	if err != nil {
		*s.errors = append(*s.errors, err.WithSourcePointer(s.ResolveReference(reference)).WithMeta(s.meta))
	}
}

func (s *Standard) NewChildContext(reference interface{}) data.Context {
	return &Standard{
		logger:  s.logger,
		pointer: s.ResolveReference(reference),
		meta:    s.meta,
		errors:  s.errors,
	}
}
