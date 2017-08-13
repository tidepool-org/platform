package test

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
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

	if len(a.InitializeMiddlewareOutputs) == 0 {
		panic("Unexpected invocation of InitializeMiddleware on API")
	}

	output := a.InitializeMiddlewareOutputs[0]
	a.InitializeMiddlewareOutputs = a.InitializeMiddlewareOutputs[1:]
	return output
}

func (a *API) InitializeRouter(routes ...*rest.Route) error {
	a.InitializeRouterInvocations++

	a.InitializeRouterInputs = append(a.InitializeRouterInputs, routes)

	if len(a.InitializeRouterOutputs) == 0 {
		panic("Unexpected invocation of InitializeRouter on API")
	}

	output := a.InitializeRouterOutputs[0]
	a.InitializeRouterOutputs = a.InitializeRouterOutputs[1:]
	return output
}

func (a *API) Status() *rest.Status {
	a.StatusInvocations++

	if len(a.StatusOutputs) == 0 {
		panic("Unexpected invocation of Status on API")
	}

	output := a.StatusOutputs[0]
	a.StatusOutputs = a.StatusOutputs[1:]
	return output
}

func (a *API) Handler() http.Handler {
	a.HandlerInvocations++

	if len(a.HandlerOutputs) == 0 {
		panic("Unexpected invocation of Handler on API")
	}

	output := a.HandlerOutputs[0]
	a.HandlerOutputs = a.HandlerOutputs[1:]
	return output
}

func (a *API) UnusedOutputsCount() int {
	return len(a.InitializeMiddlewareOutputs) +
		len(a.InitializeRouterOutputs) +
		len(a.StatusOutputs) +
		len(a.HandlerOutputs)
}
