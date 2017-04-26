package client

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/userservices/client"
)

type Context interface {
	Logger() log.Logger
	Request() *rest.Request
	UserServicesClient() client.Client
}

type Client interface {
	DestroyDataForUserByID(context Context, userID string) error
}
