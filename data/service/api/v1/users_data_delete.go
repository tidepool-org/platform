package v1

import (
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
)

func UsersDataDelete(dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	lgr := log.LoggerFromContext(ctx)

	targetUserID := dataServiceContext.Request().PathParam("userId")
	if targetUserID == "" {
		dataServiceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	if details := request.GetAuthDetails(ctx); !details.IsService() {
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

	if err := dataServiceContext.MetricClient().RecordMetric(ctx, "users_data_delete"); err != nil {
		lgr.WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, []struct{}{})
}
