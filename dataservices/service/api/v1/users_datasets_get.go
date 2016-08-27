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
	commonService "github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/userservices/client"
)

func UsersDatasetsGet(serviceContext service.Context) {
	targetUserID := serviceContext.Request().PathParam("userid")
	if targetUserID == "" {
		serviceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	if !serviceContext.IsAuthenticatedServer() {
		permissions, err := serviceContext.UserServicesClient().GetUserPermissions(serviceContext, serviceContext.AuthenticatedUserID(), targetUserID)
		if err != nil {
			if client.IsUnauthorizedError(err) {
				serviceContext.RespondWithError(commonService.ErrorUnauthorized())
			} else {
				serviceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[client.ViewPermission]; !ok {
			serviceContext.RespondWithError(commonService.ErrorUnauthorized())
			return
		}
	}

	datasets, err := serviceContext.DataStoreSession().GetDatasetsForUser(targetUserID)
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to get datasets for user", err)
		return
	}

	serviceContext.RespondWithStatusAndData(http.StatusOK, datasets)
}
