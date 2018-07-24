package v1

import (
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
)

func DataSetsDelete(dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	lgr := log.LoggerFromContext(ctx)

	dataSetID := dataServiceContext.Request().PathParam("dataSetId")
	if dataSetID == "" {
		dataServiceContext.RespondWithError(ErrorDataSetIDMissing())
		return
	}

	dataSet, err := dataServiceContext.DataSession().GetDataSetByID(ctx, dataSetID)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get data set by id", err)
		return
	}
	if dataSet == nil {
		dataServiceContext.RespondWithError(ErrorDataSetIDNotFound(dataSetID))
		return
	}

	targetUserID := dataSet.UserID
	if targetUserID == nil || *targetUserID == "" {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get user id from data set")
		return
	}

	if details := request.DetailsFromContext(ctx); !details.IsService() {
		authUserID := details.UserID()

		var permissions permission.Permissions
		permissions, err = dataServiceContext.PermissionClient().GetUserPermissions(ctx, authUserID, *targetUserID)
		if err != nil {
			if request.IsErrorUnauthorized(err) {
				dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			} else {
				dataServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[permission.Owner]; !ok {
			if _, ok = permissions[permission.Custodian]; !ok {
				if _, ok = permissions[permission.Write]; !ok || dataSet.ByUser == nil || authUserID != *dataSet.ByUser {
					dataServiceContext.RespondWithError(service.ErrorUnauthorized())
					return
				}
			}
		}
	}

	registered, err := dataServiceContext.DataDeduplicatorFactory().IsRegisteredWithDataSet(dataSet)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to check if registered with data set", err)
		return
	}

	if registered {
		deduplicator, newErr := dataServiceContext.DataDeduplicatorFactory().NewRegisteredDeduplicatorForDataSet(lgr, dataServiceContext.DataSession(), dataSet)
		if newErr != nil {
			dataServiceContext.RespondWithInternalServerFailure("Unable to create registered deduplicator for data set", newErr)
			return
		}
		err = deduplicator.DeleteDataSet(ctx)
	} else {
		err = dataServiceContext.DataSession().DeleteDataSet(ctx, dataSet)
	}

	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to delete data set", err)
		return
	}

	if err = dataServiceContext.MetricClient().RecordMetric(ctx, "data_sets_delete"); err != nil {
		lgr.WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, struct{}{})
}
