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
	"github.com/tidepool-org/platform/dataservices/server/api/v1/errors"
	"github.com/tidepool-org/platform/userservices/client"
)

func Authenticate(handler api.HandlerFunc) api.HandlerFunc {
	return func(context *api.Context) {
		userSessionToken := context.Request().Header.Get(client.TidepoolUserSessionTokenHeaderName)
		if userSessionToken == "" {
			context.RespondWithError(errors.ConstructError(errors.AuthenticationTokenMissing))
			return
		}

		requestUserID, err := context.Client().ValidateUserSession(context.Context, userSessionToken)
		if err != nil {
			if client.IsUnauthorizedError(err) {
				context.RespondWithError(errors.ConstructError(errors.Unauthenticated))
			} else {
				context.RespondWithServerFailure("Unable to validate user session", err, userSessionToken)
			}
			return
		}

		context.RequestUserID = requestUserID

		handler(context)
	}
}
