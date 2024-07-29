package mocks

import (
	stdcontext "context"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/data/service/context"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	servicecontext "github.com/tidepool-org/platform/service/context"
	"github.com/tidepool-org/platform/service/test"
)

// Context is a mock of context.Standard.
type Context struct {
	*context.Standard

	T likeT
	// authDetails should be updated via the WithAuthDetails method.
	authDetails          *test.MockAuthDetails
	RESTRequest          *rest.Request
	ResponseWriter       rest.ResponseWriter
	recorder             *httptest.ResponseRecorder
	MockAlertsRepository alerts.Repository
	MockPermissionClient permission.Client
}

func NewContext(t likeT, method, url string, body io.Reader) *Context {
	details := DefaultAuthDetails()
	ctx := request.NewContextWithAuthDetails(stdcontext.Background(), details)
	r, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		t.Fatalf("error creating request: %s", err)
	}

	w := test.NewMockRestResponseWriter()

	rr := &rest.Request{
		Request:    r,
		PathParams: map[string]string{"followerUserId": TestUserID1, "userId": TestUserID2},
		Env:        map[string]interface{}{},
	}
	responder, err := servicecontext.NewResponder(w, rr)
	if err != nil {
		t.Fatalf("error creating responder: %s", err)
	}

	return &Context{
		authDetails: details,
		Standard: &context.Standard{
			Responder: responder,
		},
		RESTRequest:          rr,
		ResponseWriter:       w,
		MockPermissionClient: NewPermission(TestPerms(), nil, nil),
		recorder:             w.ResponseRecorder,
		T:                    t,
	}
}

func (c *Context) WithAuthDetails(authDetails *test.MockAuthDetails) {
	c.authDetails = authDetails
	r := c.RESTRequest.Request
	ctx := request.NewContextWithAuthDetails(r.Context(), authDetails)
	c.RESTRequest.Request = r.WithContext(ctx)
}

// DefaultAuthDetails provides details for TestUser #1.
func DefaultAuthDetails() *test.MockAuthDetails {
	return test.NewMockAuthDetails(request.MethodSessionToken, test.TestUserID1, test.TestToken1)
}

// ServiceAuthDetails provides details for a service call.
func ServiceAuthDetails() *test.MockAuthDetails {
	return test.NewMockAuthDetails(request.MethodServiceSecret, "", test.TestToken2)
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

func (c *Context) PermissionClient() permission.Client {
	return c.MockPermissionClient
}
