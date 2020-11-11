package service

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
)

type API interface {
	InitializeMiddleware(name string) error
	InitializeRouters(routers ...Router) error

	Status() *rest.Status

	Handler() http.Handler
}

type Router interface {
	Routes() []*rest.Route
}
