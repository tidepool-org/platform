package v1

import (
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
)

// UsersDataDelete godoc
// @Summary Delete user data
// @Description Only services (eg. not users) can delete user data
// @ID platform-data-api-UsersDataDelete
// @Accept json
// @Produce json
// @Param userId path string true "user ID"
// @Security TidepoolSessionToken
// @Security TidepoolServiceSecret
// @Success 200 {object} upload.Upload "Operation is a success"
// @Failure 400 {object} service.Error "User id is missing"
// @Failure 403 {object} service.Error "Forbiden: caller is not a service"
// @Failure 500 {object} service.Error "Unable to perform the operation"
// @Router /v1/users/:userId/data [delete]
func UsersDataDelete(dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()

	targetUserID := dataServiceContext.Request().PathParam("userId")
	if targetUserID == "" {
		dataServiceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	if details := request.DetailsFromContext(ctx); !details.IsService() {
		dataServiceContext.RespondWithError(service.ErrorUnauthorized())
		return
	}

	if err := dataServiceContext.DataRepository().DestroyDataForUserByID(ctx, targetUserID); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to delete data for user by id", err)
		return
	}

	// TODO: This should probably be in its own API, but then again, these are very specific sync tasks and
	// the whole sync task thing needs to be reworked, so we'll leave it be for the time being.
	if err := dataServiceContext.SyncTaskRepository().DestroySyncTasksForUserByID(ctx, targetUserID); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to delete sync tasks for user by id", err)
		return
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, []struct{}{})
}
