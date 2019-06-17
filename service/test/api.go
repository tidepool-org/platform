package test

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/service"
)

type API struct {
	InitializeMiddlewareInvocations int
	InitializeMiddlewareOutputs     []error
	InitializeRoutersInvocations    int
	InitializeRoutersInputs         [][]service.Router
	InitializeRoutersOutputs        []error
	StatusInvocations               int
	StatusOutputs                   []*rest.Status
	HandlerInvocations              int
	HandlerOutputs                  []http.Handler
}

func NewAPI() *API {
	return &API{}
}

func (a *API) InitializeMiddleware() error {
	a.InitializeMiddlewareInvocations++

	gomega.Expect(a.InitializeMiddlewareOutputs).ToNot(gomega.BeEmpty())

	output := a.InitializeMiddlewareOutputs[0]
	a.InitializeMiddlewareOutputs = a.InitializeMiddlewareOutputs[1:]
	return output
}

func (a *API) InitializeRouters(routes ...service.Router) error {
	a.InitializeRoutersInvocations++

	a.InitializeRoutersInputs = append(a.InitializeRoutersInputs, routes)

	gomega.Expect(a.InitializeRoutersOutputs).ToNot(gomega.BeEmpty())

	output := a.InitializeRoutersOutputs[0]
	a.InitializeRoutersOutputs = a.InitializeRoutersOutputs[1:]
	return output
}

func (a *API) Status() *rest.Status {
	a.StatusInvocations++

	gomega.Expect(a.StatusOutputs).ToNot(gomega.BeEmpty())

	output := a.StatusOutputs[0]
	a.StatusOutputs = a.StatusOutputs[1:]
	return output
}

func (a *API) Handler() http.Handler {
	a.HandlerInvocations++

	gomega.Expect(a.HandlerOutputs).ToNot(gomega.BeEmpty())

	output := a.HandlerOutputs[0]
	a.HandlerOutputs = a.HandlerOutputs[1:]
	return output
}

func (a *API) Expectations() {
	gomega.Expect(a.InitializeMiddlewareOutputs).To(gomega.BeEmpty())
	gomega.Expect(a.InitializeRoutersOutputs).To(gomega.BeEmpty())
	gomega.Expect(a.StatusOutputs).To(gomega.BeEmpty())
	gomega.Expect(a.HandlerOutputs).To(gomega.BeEmpty())
}
