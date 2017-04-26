package v1

import (
	"net/http"

	"github.com/tidepool-org/platform/dataservices/service"
	commonService "github.com/tidepool-org/platform/service"
)

func UsersDataDelete(serviceContext service.Context) {
	targetUserID := serviceContext.Request().PathParam("userid")
	if targetUserID == "" {
		serviceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	if !serviceContext.AuthenticationDetails().IsServer() {
		serviceContext.RespondWithError(commonService.ErrorUnauthorized())
		return
	}

	if err := serviceContext.DataStoreSession().DestroyDataForUserByID(targetUserID); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to delete data for user by id", err)
		return
	}

	// TODO: This should probably be in its own API, but then again, these are very specific tasks and
	// the whole syncTask thing needs to be reworked, so we'll leave it be for the time being.
	if err := serviceContext.TaskStoreSession().DestroyTasksForUserByID(targetUserID); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to delete tasks for user by id", err)
		return
	}

	if err := serviceContext.MetricServicesClient().RecordMetric(serviceContext, "users_data_delete"); err != nil {
		serviceContext.Logger().WithError(err).Error("Unable to record metric")
	}

	serviceContext.RespondWithStatusAndData(http.StatusOK, []struct{}{})
}
