package test

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/test"
)

type API struct {
	*test.Mock
	InitializeMiddlewareInvocations int
	InitializeMiddlewareOutputs     []error
	InitializeRouterInvocations     int
	InitializeRouterInputs          [][]*rest.Route
	InitializeRouterOutputs         []error
	StatusInvocations               int
	StatusOutputs                   []*rest.Status
	HandlerInvocations              int
	HandlerOutputs                  []http.Handler
}

func NewAPI() *API {
	return &API{
		Mock: test.NewMock(),
	}
}

func (a *API) InitializeMiddleware() error {
	a.InitializeMiddlewareInvocations++

	gomega.Expect(a.InitializeMiddlewareOutputs).ToNot(gomega.BeEmpty())

	output := a.InitializeMiddlewareOutputs[0]
	a.InitializeMiddlewareOutputs = a.InitializeMiddlewareOutputs[1:]
	return output
}

func (a *API) InitializeRouter(routes ...*rest.Route) error {
	a.InitializeRouterInvocations++

	a.InitializeRouterInputs = append(a.InitializeRouterInputs, routes)

	gomega.Expect(a.InitializeRouterOutputs).ToNot(gomega.BeEmpty())

	output := a.InitializeRouterOutputs[0]
	a.InitializeRouterOutputs = a.InitializeRouterOutputs[1:]
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
	a.Mock.Expectations()
	gomega.Expect(a.InitializeMiddlewareOutputs).To(gomega.BeEmpty())
	gomega.Expect(a.InitializeRouterOutputs).To(gomega.BeEmpty())
	gomega.Expect(a.StatusOutputs).To(gomega.BeEmpty())
	gomega.Expect(a.HandlerOutputs).To(gomega.BeEmpty())
}
