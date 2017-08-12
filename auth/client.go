package auth

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/log"
)

type Client interface {
	ServerToken() (string, error)

	ValidateToken(context Context, token string) (Details, error)
}

type Context interface {
	Logger() log.Logger
	Request() *rest.Request

	AuthClient() Client
	AuthDetails() Details
}

type Details interface {
	Token() string

	IsServer() bool
	UserID() string
}

const TidepoolAuthTokenHeaderName = "X-Tidepool-Session-Token"
