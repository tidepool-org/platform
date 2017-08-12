package rest

import (
	"github.com/onsi/gomega"

	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
)

func NewRequest() *rest.Request {
	request, err := http.NewRequest("GET", "http://127.0.0.1/", nil)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(request).ToNot(gomega.BeNil())
	return &rest.Request{
		Request:    request,
		PathParams: map[string]string{},
		Env:        map[string]interface{}{},
	}
}
