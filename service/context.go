package service

import (
	"github.com/mdblp/go-json-rest/rest"
)

type Context interface {
	Response() rest.ResponseWriter
	Request() *rest.Request

	RespondWithError(err *Error)
	RespondWithInternalServerFailure(message string, failure ...interface{})
	RespondWithStatusAndErrors(statusCode int, errors []*Error)
	RespondWithStatusAndData(statusCode int, data interface{})
}

type Meta struct {
	Trace *Trace `json:"trace,omitempty"`
}

type Trace struct {
	Request string `json:"request,omitempty"`
	Session string `json:"session,omitempty"`
}
