package v1

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
	"net/http"

	"github.com/tidepool-org/platform/dataservices/server/api"
	"github.com/tidepool-org/platform/userservices/client"
)

func UsersCheck(context *api.Context) {
	targetUserID := context.Request().PathParam("userid")
	if targetUserID == "" {
		context.RespondWithError(ErrorUserIDMissing())
		return
	}

	err := context.Client().ValidateTargetUserPermissions(context.Context, context.RequestUserID, targetUserID, client.UploadPermissions)
	if err != nil {
		if client.IsUnauthorizedError(err) {
			context.RespondWithError(ErrorUnauthorized())
		} else {
			context.RespondWithServerFailure("Unable to validate target user permissions", err)
		}
		return
	}

	targetUserGroupID, err := context.Client().GetUserGroupID(context.Context, targetUserID)
	if err != nil {
		if client.IsUnauthorizedError(err) {
			context.RespondWithError(ErrorUnauthorized())
		} else {
			context.RespondWithServerFailure("Unable to get group id for user", err)
		}
		return
	}

	context.Response().WriteHeader(http.StatusOK)
	context.Response().WriteJson(map[string]string{"requestUserID": context.RequestUserID, "targetUserID": targetUserID, "targetGroupID": targetUserGroupID})
}
