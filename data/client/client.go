package client

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/log"
	userClient "github.com/tidepool-org/platform/user/client"
)

type Context interface {
	Logger() log.Logger
	Request() *rest.Request
	UserClient() userClient.Client
}

type Client interface {
	DestroyDataForUserByID(context Context, userID string) error
}
