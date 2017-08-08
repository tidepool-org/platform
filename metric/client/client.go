package client

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/log"
	userClient "github.com/tidepool-org/platform/user/client"
)

type Context interface {
	Logger() log.Logger
	Request() *rest.Request
	AuthenticationDetails() userClient.AuthenticationDetails
}

type Client interface {
	RecordMetric(context Context, name string, data ...map[string]string) error
}
