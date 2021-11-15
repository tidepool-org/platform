package service

import (
	"net/http"

	"github.com/mdblp/go-json-rest/rest"
)

type API interface {
	InitializeMiddleware() error
	InitializeRouters(routers ...Router) error

	Status() *rest.Status

	Handler() http.Handler
}

type Router interface {
	Routes() []*rest.Route
}
