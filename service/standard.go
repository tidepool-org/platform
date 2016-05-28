package service

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

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/log"
)

type Standard struct {
	response rest.ResponseWriter
	request  *rest.Request
	logger   log.Logger
}

type Trace struct {
	Request string `json:"request,omitempty"`
	Session string `json:"session,omitempty"`
}

type Meta struct {
	Trace *Trace `json:"trace,omitempty"`
}

type JSONResponse struct {
	Errors []*Error `json:"errors,omitempty"`
	Meta   *Meta    `json:"meta,omitempty"`
}

func NewStandard(response rest.ResponseWriter, request *rest.Request) *Standard {
	return &Standard{
		response: response,
		request:  request,
		logger:   GetRequestLogger(request),
	}
}

func (s *Standard) Logger() log.Logger {
	return s.logger
}

func (s *Standard) Request() *rest.Request {
	return s.request
}

func (s *Standard) Response() rest.ResponseWriter {
	return s.response
}

func (s *Standard) RespondWithError(err *Error) {
	if statusCode := err.Status; statusCode <= 0 { // TODO: Do we want to validate the status code is okay? More than >= 0?
		s.RespondWithInternalServerFailure("Status field missing from error", err)
	} else {
		s.RespondWithStatusAndErrors(statusCode, []*Error{err})
	}
}

func (s *Standard) RespondWithInternalServerFailure(message string, failure ...interface{}) {
	logger := s.Logger()
	if len(failure) > 0 {
		for index := range failure {
			if err, errOk := failure[index].(error); errOk {
				failure[index] = err.Error()
			} else if stringer, stringerOk := failure[index].(fmt.Stringer); stringerOk {
				failure[index] = stringer.String()
			}
		}
		logger = logger.WithField("failure", failure)
	}
	logger.Error(message)
	s.RespondWithError(ErrorInternalServerFailure())
}

func (s *Standard) RespondWithStatusAndErrors(statusCode int, errors []*Error) {
	s.Request().Env[RequestEnvErrors] = errors

	response := &JSONResponse{
		Errors: errors,
		Meta: &Meta{
			Trace: &Trace{
				Request: GetRequestTraceRequest(s.Request()),
			},
		},
	}

	if traceSession := GetRequestTraceSession(s.Request()); traceSession != "" {
		response.Meta.Trace.Session = traceSession
	}

	s.Response().WriteHeader(statusCode)
	s.Response().WriteJson(response)
}
