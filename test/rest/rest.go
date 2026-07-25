package rest

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	testHttp "github.com/tidepool-org/platform/test/http"
)

//go:generate mockgen -source=rest.go -destination=rest_mocks.go -package=rest -typed

type Response interface {
	http.ResponseWriter
	rest.ResponseWriter
}

func NewRequest() *rest.Request {
	return &rest.Request{
		Request:    testHttp.NewRequest(),
		PathParams: map[string]string{},
		Env:        map[string]interface{}{},
	}
}
