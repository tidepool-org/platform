package context

import (
	"fmt"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
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

type Responder struct {
	response rest.ResponseWriter
	request  *rest.Request
	logger   log.Logger
}

func NewResponder(response rest.ResponseWriter, request *rest.Request) (*Responder, error) {
	if response == nil {
		return nil, errors.New("response is missing")
	}
	if request == nil {
		return nil, errors.New("request is missing")
	}

	return &Responder{
		response: response,
		request:  request,
	}, nil
}

func (r *Responder) Response() rest.ResponseWriter {
	return r.response
}

func (r *Responder) Request() *rest.Request {
	return r.request
}

func (r *Responder) Logger() log.Logger {
	if r.logger == nil {
		logger := service.GetRequestLogger(r.Request())
		if logger == nil {
			logger = null.NewLogger()
		}
		r.logger = logger
	}

	return r.logger
}

func (r *Responder) RespondWithError(err *service.Error) {
	if err == nil {
		r.RespondWithInternalServerFailure("Error is missing")
	} else if err.Status <= 0 {
		r.RespondWithInternalServerFailure("Status field is missing from error", err)
	} else {
		r.RespondWithStatusAndErrors(err.Status, []*service.Error{err})
	}
}

func (r *Responder) RespondWithInternalServerFailure(message string, failure ...interface{}) {
	logger := r.Logger()
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
	r.RespondWithError(service.ErrorInternalServerFailure())
}

func (r *Responder) RespondWithStatusAndErrors(statusCode int, errors []*service.Error) {
	service.SetRequestErrors(r.Request(), errors)

	response := &JSONResponse{
		Errors: errors,
		Meta: &Meta{
			Trace: &Trace{
				Request: service.GetRequestTraceRequest(r.Request()),
				Session: service.GetRequestTraceSession(r.Request()),
			},
		},
	}

	r.Response().WriteHeader(statusCode)
	r.Response().WriteJson(response)
}

func (r *Responder) RespondWithStatusAndData(statusCode int, data interface{}) {
	response := &JSONResponse{
		Data: data,
		Meta: &Meta{
			Trace: &Trace{
				Request: service.GetRequestTraceRequest(r.Request()),
				Session: service.GetRequestTraceSession(r.Request()),
			},
		},
	}

	r.Response().WriteHeader(statusCode)
	r.Response().WriteJson(response)
}
