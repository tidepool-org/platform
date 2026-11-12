package v1

import (
	"github.com/ant0ine/go-json-rest/rest"
)

type Context any

type Route struct {
	Method     string
	Path       string
	Handler    func(Context)
	Middleware []rest.MiddlewareSimple
}

func Routes() []Route {
	return nil
}
