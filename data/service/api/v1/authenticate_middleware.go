package v1

import (
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/user/client"
)

func Authenticate(handler dataService.HandlerFunc) dataService.HandlerFunc {
	return func(context dataService.Context) {
		authenticationToken := context.Request().Header.Get(client.TidepoolAuthenticationTokenHeaderName)
		if authenticationToken == "" {
			context.RespondWithError(service.ErrorAuthenticationTokenMissing())
			return
		}

		authenticationDetails, err := context.UserClient().ValidateAuthenticationToken(context, authenticationToken)
		if err != nil {
			if client.IsUnauthorizedError(err) {
				context.RespondWithError(service.ErrorUnauthenticated())
			} else {
				context.RespondWithInternalServerFailure("Unable to validate authentication token", err, authenticationToken)
			}
			return
		}

		context.SetAuthenticationDetails(authenticationDetails)

		handler(context)
	}
}
