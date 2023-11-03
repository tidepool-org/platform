package service

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
)

type Route struct {
	Handler    HandlerFunc
	Method     string
	Path       string
	middleware []rest.MiddlewareSimple
}

// MakeRoute builds a Route.
//
// Consider using the handy Get, Post, etc helpers.
func MakeRoute(method string, path string, handler HandlerFunc, middleware ...rest.MiddlewareSimple) Route {
	return Route{
		Method:     method,
		Path:       path,
		Handler:    handler,
		middleware: middleware,
	}
}

// Delete wraps MakeRoute for easy DELETE route creation.
func Delete(path string, handler HandlerFunc, middleware ...rest.MiddlewareSimple) Route {
	return MakeRoute(http.MethodDelete, path, handler, middleware...)
}

// Get wraps MakeRoute for easy GET route creation.
func Get(path string, handler HandlerFunc, middleware ...rest.MiddlewareSimple) Route {
	return MakeRoute(http.MethodGet, path, handler, middleware...)
}

// Patch wraps MakeRoute for easy PATCH route creation.
func Patch(path string, handler HandlerFunc, middleware ...rest.MiddlewareSimple) Route {
	return MakeRoute(http.MethodPatch, path, handler, middleware...)
}

// Post wraps MakeRoute for easy POST route creation.
func Post(path string, handler HandlerFunc, middleware ...rest.MiddlewareSimple) Route {
	return MakeRoute(http.MethodPost, path, handler, middleware...)
}

// Put wraps MakeRoute for easy PUT route creation.
func Put(path string, handler HandlerFunc, middleware ...rest.MiddlewareSimple) Route {
	return MakeRoute(http.MethodPut, path, handler, middleware...)
}

// RestRouteAdapterFunc adapts a HandlerFunc to a rest.HandlerFunc.
type RestRouteAdapterFunc (func(HandlerFunc) rest.HandlerFunc)

// ToRestRoute converts a Route to a rest.Route.
func (r *Route) ToRestRoute(f RestRouteAdapterFunc) *rest.Route {
	var middlewares []rest.Middleware
	for _, s := range r.middleware {
		middlewares = append(middlewares, rest.MiddlewareSimple(s))
	}
	return &rest.Route{
		HttpMethod: r.Method,
		PathExp:    r.Path,
		Func:       rest.WrapMiddlewares(middlewares, f(r.Handler)),
	}
}
