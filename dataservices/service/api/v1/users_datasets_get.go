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

	"github.com/tidepool-org/platform/dataservices/service"
	"github.com/tidepool-org/platform/userservices/client"
)

func UsersDatasetsGet(serviceContext service.Context) {
	targetUserID := serviceContext.Request().PathParam("userid")
	if targetUserID == "" {
		serviceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	permissions, err := serviceContext.UserServicesClient().GetUserPermissions(serviceContext, serviceContext.RequestUserID(), targetUserID)
	if err != nil {
		if client.IsUnauthorizedError(err) {
			serviceContext.RespondWithError(ErrorUnauthorized())
		} else {
			serviceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
		}
		return
	}
	if _, ok := permissions[client.ViewPermission]; !ok {
		serviceContext.RespondWithError(ErrorUnauthorized())
		return
	}

	datasets, err := serviceContext.DataStoreSession().GetDatasetsForUser(targetUserID)
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to get datasets for user", err)
		return
	}

	serviceContext.RespondWithStatusAndData(http.StatusOK, datasets)
}
