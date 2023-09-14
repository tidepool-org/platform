package mocks

import (
	stdcontext "context"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/data/service/context"
	"github.com/tidepool-org/platform/request"
	servicecontext "github.com/tidepool-org/platform/service/context"
)

// Context is a mock of context.Standard.
type Context struct {
	*context.Standard

	T likeT
	// details should be updated via the WithDetails method.
	details              *Details
	RESTRequest          *rest.Request
	ResponseWriter       rest.ResponseWriter
	recorder             *httptest.ResponseRecorder
	MockAlertsRepository alerts.Repository
}

func NewContext(t likeT, method, url string, body io.Reader) *Context {
	details := defDetails()
	ctx := request.NewContextWithDetails(stdcontext.Background(), details)
	r, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		t.Fatalf("error creating request: %s", err)
	}

	recorder := httptest.NewRecorder()
	w := NewResponseWriter(recorder)

	rr := &rest.Request{
		Request:    r,
		PathParams: map[string]string{},
		Env:        map[string]interface{}{},
	}
	responder, err := servicecontext.NewResponder(w, rr)
	if err != nil {
		t.Fatalf("error creating responder: %s", err)
	}

	return &Context{
		details: details,
		Standard: &context.Standard{
			Responder: responder,
		},
		RESTRequest:    rr,
		ResponseWriter: w,
		recorder:       recorder,
		T:              t,
	}
}

func (c *Context) WithDetails(details *Details) {
	c.details = details
	r := c.RESTRequest.Request
	ctx := request.NewContextWithDetails(r.Context(), details)
	c.RESTRequest.Request = r.WithContext(ctx)
}

func defDetails() *Details {
	return NewDetails(request.MethodSessionToken, TestUserID1, TestToken1)
}

func (c *Context) Response() rest.ResponseWriter {
	return c.ResponseWriter
}

func (c *Context) Request() *rest.Request {
	return c.RESTRequest
}

func (c *Context) Recorder() *httptest.ResponseRecorder {
	return c.recorder
}

func (c *Context) AlertsRepository() alerts.Repository {
	return c.MockAlertsRepository
}
