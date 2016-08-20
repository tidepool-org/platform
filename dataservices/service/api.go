package service

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import "net/http"

type API interface {
	Handler() http.Handler
}

type Route struct {
	Method  string
	Path    string
	Handler HandlerFunc
}

func MakeRoute(method string, path string, handler HandlerFunc) Route {
	return Route{
		Method:  method,
		Path:    path,
		Handler: handler,
	}
}
