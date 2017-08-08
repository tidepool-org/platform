package v1

import (
	"github.com/tidepool-org/platform/service"
	userClient "github.com/tidepool-org/platform/user/client"
	userService "github.com/tidepool-org/platform/user/service"
)

func Authenticate(handler userService.HandlerFunc) userService.HandlerFunc {
	return func(context userService.Context) {
		authenticationToken := context.Request().Header.Get(userClient.TidepoolAuthenticationTokenHeaderName)
		if authenticationToken == "" {
			context.RespondWithError(service.ErrorAuthenticationTokenMissing())
			return
		}

		authenticationDetails, err := context.UserClient().ValidateAuthenticationToken(context, authenticationToken)
		if err != nil {
			if userClient.IsUnauthorizedError(err) {
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
