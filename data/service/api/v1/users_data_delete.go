package v1

import (
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/service"
)

func UsersDataDelete(dataServiceContext dataService.Context) {
	targetUserID := dataServiceContext.Request().PathParam("userid")
	if targetUserID == "" {
		dataServiceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	if !dataServiceContext.AuthenticationDetails().IsServer() {
		dataServiceContext.RespondWithError(service.ErrorUnauthorized())
		return
	}

	if err := dataServiceContext.DataStoreSession().DestroyDataForUserByID(targetUserID); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to delete data for user by id", err)
		return
	}

	// TODO: This should probably be in its own API, but then again, these are very specific sync tasks and
	// the whole sync task thing needs to be reworked, so we'll leave it be for the time being.
	if err := dataServiceContext.SyncTaskStoreSession().DestroySyncTasksForUserByID(targetUserID); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to delete sync tasks for user by id", err)
		return
	}

	if err := dataServiceContext.MetricClient().RecordMetric(dataServiceContext, "users_data_delete"); err != nil {
		dataServiceContext.Logger().WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, []struct{}{})
}
