package context

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

	commonService "github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/context"
	"github.com/tidepool-org/platform/userservices/client"
	"github.com/tidepool-org/platform/userservices/service"
)

type Standard struct {
	commonService.Context
	userServicesClient    client.Client
	authenticationDetails client.AuthenticationDetails
}

func WithContext(userServicesClient client.Client, handler service.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		context, err := context.NewStandard(response, request)
		if err != nil {
			context.RespondWithInternalServerFailure("Unable to create new context for request", err)
			return
		}

		handler(&Standard{
			Context:            context,
			userServicesClient: userServicesClient,
		})
	}
}

func (s *Standard) UserServicesClient() client.Client {
	return s.userServicesClient
}

func (s *Standard) SetAuthenticationDetails(authenticationDetails client.AuthenticationDetails) {
	s.authenticationDetails = authenticationDetails
}

func (s *Standard) IsAuthenticatedServer() bool {
	if s.authenticationDetails == nil {
		return false
	}
	return s.authenticationDetails.IsServer()
}

func (s *Standard) AuthenticatedUserID() string {
	if s.authenticationDetails == nil {
		return ""
	}
	return s.authenticationDetails.UserID()
}
