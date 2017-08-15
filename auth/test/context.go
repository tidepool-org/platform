package test

import (
	"net/http"
	"net/url"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/log"
	nullLog "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/test"
)

type Context struct {
	*test.Mock
	LoggerImpl      log.Logger
	RequestImpl     *rest.Request
	AuthClientMock  auth.Client
	AuthClientImpl  *Client
	AuthDetailsMock auth.Details
	AuthDetailsImpl *Details
}

func NewContext() *Context {
	return &Context{
		Mock:       test.NewMock(),
		LoggerImpl: nullLog.NewLogger(),
		RequestImpl: &rest.Request{
			Request: &http.Request{
				URL: &url.URL{},
			},
			PathParams: map[string]string{},
		},
		AuthClientImpl:  NewClient(),
		AuthDetailsImpl: NewDetails(),
	}
}

func (c *Context) Logger() log.Logger {
	return c.LoggerImpl
}

func (c *Context) Request() *rest.Request {
	return c.RequestImpl
}

func (c *Context) AuthClient() auth.Client {
	if c.AuthClientMock != nil {
		return c.AuthClientMock
	}
	return c.AuthClientImpl
}

func (c *Context) AuthDetails() auth.Details {
	if c.AuthDetailsMock != nil {
		return c.AuthDetailsMock
	}
	return c.AuthDetailsImpl
}

func (c *Context) UnusedOutputsCount() int {
	return c.AuthClientImpl.UnusedOutputsCount() +
		c.AuthDetailsImpl.UnusedOutputsCount()
}
