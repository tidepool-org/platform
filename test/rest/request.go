package rest

import (
	"github.com/mdblp/go-json-rest/rest"

	testHttp "github.com/tidepool-org/platform/test/http"
)

func NewRequest() *rest.Request {
	return &rest.Request{
		Request:    testHttp.NewRequest(),
		PathParams: map[string]string{},
		Env:        map[string]interface{}{},
	}
}
