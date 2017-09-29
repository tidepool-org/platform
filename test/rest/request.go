package rest

import (
	"github.com/ant0ine/go-json-rest/rest"

	testHTTP "github.com/tidepool-org/platform/test/http"
)

func NewRequest() *rest.Request {
	return &rest.Request{
		Request:    testHTTP.NewRequest(),
		PathParams: map[string]string{},
		Env:        map[string]interface{}{},
	}
}
