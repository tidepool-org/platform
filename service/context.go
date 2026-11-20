package service

import (
	"github.com/ant0ine/go-json-rest/rest"
)

type Context interface {
	Response() rest.ResponseWriter
	Request() *rest.Request

	RespondWithError(err *Error)
	RespondWithInternalServerFailure(message string, err error)
	RespondWithStatusAndData(statusCode int, data any)
}

type Meta struct {
	Trace *Trace `json:"trace,omitempty"`
}

type Trace struct {
	Request string `json:"request,omitempty"`
	Session string `json:"session,omitempty"`
}
