package middleware

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
	"github.com/tidepool-org/platform/dataservices/server/api"
	"github.com/tidepool-org/platform/dataservices/server/api/v1"
	"github.com/tidepool-org/platform/userservices/client"
)

func Authenticate(handler api.HandlerFunc) api.HandlerFunc {
	return func(context *api.Context) {
		userSessionToken := context.Request().Header.Get(client.TidepoolUserSessionTokenHeaderName)
		if userSessionToken == "" {
			context.RespondWithError(v1.ErrorAuthenticationTokenMissing())
			return
		}

		requestUserID, err := context.Client().ValidateUserSession(context.Context, userSessionToken)
		if err != nil {
			if client.IsUnauthorizedError(err) {
				context.RespondWithError(v1.ErrorUnauthenticated())
			} else {
				context.RespondWithInternalServerFailure("Unable to validate user session", err, userSessionToken)
			}
			return
		}

		context.RequestUserID = requestUserID

		handler(context)
	}
}
