package auth

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/log"
)

type Client interface {
	ServerToken() (string, error)

	ValidateToken(ctx Context, token string) (Details, error)

	GetStatus(ctx Context) (*Status, error)
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

type Status struct {
	Version   string
	Server    interface{}
	AuthStore interface{}
}

const TidepoolAuthTokenHeaderName = "X-Tidepool-Session-Token"
