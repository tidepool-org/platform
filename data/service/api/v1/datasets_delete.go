package v1

import (
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
)

type dataSetsDeleteParams struct {
	Purge bool `json:"purge" default:"false"` // true to remove the dataset from the database
}

// DataSetsDelete godoc
// @Summary Delete a DataSets
// @Description Delete a DataSets
// @ID platform-data-api-DataSetsDelete
// @Accept json
// @Param dataSetID path string true "dataSet ID"
// @Param dataSetsDeleteParams body dataSetsDeleteParams false "True to really remove the dataset and associated data"
// @Security TidepoolSessionToken
// @Security TidepoolServiceSecret
// @Security TidepoolAuthorization
// @Security TidepoolRestrictedToken
// @Success 200 "Operation is a success"
// @Failure 400 {object} service.Error "Data set id is missing"
// @Failure 403 {object} service.Error "Auth token is not authorized for requested action"
// @Failure 404 {object} service.Error "Data set with specified id not found"
// @Failure 500 {object} service.Error "Unable to perform the operation"
// @Router /v1/datasets/:dataSetID [delete]
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
		// FIXME: This is a temporary fix, it should return an error.
		dataServiceContext.RespondWithStatusAndData(http.StatusOK, ErrorDataSetIDNotFound(dataSetID))
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

	// Read delete options (remove dataset entry ?):
	var jsonParams map[string]interface{}
	doPurge := false
	if err := dataServiceContext.Request().DecodeJsonPayload(&jsonParams); err != nil {
		jsonParams = nil
	} else {
		purge, havePurgeOption := jsonParams["purge"]
		doPurge = havePurgeOption && purge == true
	}

	if deduplicator, getErr := dataServiceContext.DataDeduplicatorFactory().Get(dataSet); getErr != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get deduplicator", getErr)
		return
	} else if deduplicator == nil {
		if err = dataServiceContext.DataSession().DeleteDataSet(ctx, dataSet, doPurge); err != nil {
			dataServiceContext.RespondWithInternalServerFailure("Unable to delete data set", err)
			return
		}
	} else {
		if err = deduplicator.Delete(ctx, dataServiceContext.DataSession(), dataSet, doPurge); err != nil {
			dataServiceContext.RespondWithInternalServerFailure("Unable to delete", err)
			return
		}
	}

	if err = dataServiceContext.MetricClient().RecordMetric(ctx, "data_sets_delete"); err != nil {
		lgr.WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, struct{}{})
}
