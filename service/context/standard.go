package context

import (
	"fmt"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

type JSONResponse struct {
	Data   interface{}      `json:"data,omitempty"`
	Errors []*service.Error `json:"errors,omitempty"`
	Meta   *Meta            `json:"meta,omitempty"`
}

type Meta struct {
	Trace *Trace `json:"trace,omitempty"`
}

type Trace struct {
	Request string `json:"request,omitempty"`
	Session string `json:"session,omitempty"`
}

type Standard struct {
	response rest.ResponseWriter
	request  *rest.Request
	logger   log.Logger
}

func NewStandard(response rest.ResponseWriter, request *rest.Request) (*Standard, error) {
	if response == nil {
		return nil, errors.New("context", "response is missing")
	}
	if request == nil {
		return nil, errors.New("context", "request is missing")
	}

	return &Standard{
		response: response,
		request:  request,
	}, nil
}

func (s *Standard) Logger() log.Logger {
	if s.logger == nil {
		logger := service.GetRequestLogger(s.request)
		if logger == nil {
			logger = log.NewNull()
		}
		s.logger = logger
	}

	return s.logger
}

func (s *Standard) Request() *rest.Request {
	return s.request
}

func (s *Standard) Response() rest.ResponseWriter {
	return s.response
}

func (s *Standard) RespondWithError(err *service.Error) {
	if err == nil {
		s.RespondWithInternalServerFailure("Error is missing")
	} else if err.Status <= 0 {
		s.RespondWithInternalServerFailure("Status field is missing from error", err)
	} else {
		s.RespondWithStatusAndErrors(err.Status, []*service.Error{err})
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
	s.RespondWithError(service.ErrorInternalServerFailure())
}

func (s *Standard) RespondWithStatusAndErrors(statusCode int, errors []*service.Error) {
	service.SetRequestErrors(s.request, errors)

	response := &JSONResponse{
		Errors: errors,
		Meta: &Meta{
			Trace: &Trace{
				Request: service.GetRequestTraceRequest(s.Request()),
				Session: service.GetRequestTraceSession(s.Request()),
			},
		},
	}

	s.respondWithStatusAndResponse(statusCode, response)
}

func (s *Standard) RespondWithStatusAndData(statusCode int, data interface{}) {
	response := &JSONResponse{
		Data: data,
		Meta: &Meta{
			Trace: &Trace{
				Request: service.GetRequestTraceRequest(s.Request()),
				Session: service.GetRequestTraceSession(s.Request()),
			},
		},
	}

	s.respondWithStatusAndResponse(statusCode, response)
}

func (s *Standard) respondWithStatusAndResponse(statusCode int, response *JSONResponse) {
	service.AddDateHeader(s.response)

	s.response.WriteHeader(statusCode)
	s.response.WriteJson(response)
}
