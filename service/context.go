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

// TODO: Make this an interface

type Context struct {
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

type HandlerFunc func(context *Context)

func NewContext(response rest.ResponseWriter, request *rest.Request) *Context {
	return &Context{
		response: response,
		request:  request,
		logger:   GetRequestLogger(request),
	}
}

func WithContext(handler HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		handler(NewContext(response, request))
	}
}

func (c *Context) Logger() log.Logger {
	return c.logger
}

func (c *Context) Request() *rest.Request {
	return c.request
}

func (c *Context) Response() rest.ResponseWriter {
	return c.response
}

func (c *Context) RespondWithError(err *Error) {
	if statusCode := err.Status; statusCode <= 0 { // TODO: Do we want to validate the status code is okay? More than >= 0?
		c.RespondWithServerFailure("Status field missing from error", err)
	} else {
		c.RespondWithStatusAndErrors(statusCode, []*Error{err})
	}
}

func (c *Context) RespondWithServerFailure(message string, failure ...interface{}) {
	logger := c.Logger()
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
	c.RespondWithError(InternalServerFailure)
}

func (c *Context) RespondWithStatusAndErrors(statusCode int, errors []*Error) {
	c.Request().Env[RequestEnvErrors] = errors

	response := &JSONResponse{
		Errors: errors,
		Meta: &Meta{
			Trace: &Trace{
				Request: GetRequestTraceRequest(c.Request()),
			},
		},
	}

	if traceSession := GetRequestTraceSession(c.Request()); traceSession != "" {
		response.Meta.Trace.Session = traceSession
	}

	c.Response().WriteHeader(statusCode)
	c.Response().WriteJson(response)
}
