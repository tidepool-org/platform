package client

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

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
	DeleteDataForUser(context Context, userID string) error
}
