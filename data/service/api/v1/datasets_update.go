package v1

import (
	"net/http"

	"github.com/tidepool-org/platform/data"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
)

// DataSetsUpdate godoc
// @Summary Update a data sets
// @ID platform-data-api-DataSetsUpdate
// @Accept json
// @Produce json
// @Param dataSetID path string true "dataSet ID"
// @Param dataSetUpdate body data.DataSetUpdate true "The dataSet to update"
// @Security TidepoolSessionToken
// @Security TidepoolServiceSecret
// @Security TidepoolAuthorization
// @Security TidepoolRestrictedToken
// @Success 200 {object} upload.Upload "Operation is a success"
// @Failure 400 {object} service.Error "Data set id is missing"
// @Failure 403 {object} service.Error "Auth token is not authorized for requested action"
// @Failure 404 {object} service.Error "Data set with specified id not found"
// @Failure 409 {object} service.Error "Data set with specified id is closed for new data"
// @Failure 500 {object} service.Error "Unable to perform the operation"
// @Router /v1/datasets/:dataSetId [put]
func DataSetsUpdate(dataServiceContext dataService.Context) {
	req := dataServiceContext.Request()
	res := dataServiceContext.Response()
	ctx := req.Context()

	dataSetID := req.PathParam("dataSetId")
	if dataSetID == "" {
		dataServiceContext.RespondWithError(ErrorDataSetIDMissing())
		return
	}

	dataSet, err := dataServiceContext.DataRepository().GetDataSetByID(ctx, dataSetID)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get data set by id", err)
		return
	}
	if dataSet == nil {
		dataServiceContext.RespondWithError(ErrorDataSetIDNotFound(dataSetID))
		return
	}

	permissions, err := dataServiceContext.PermissionClient().GetUserPermissions(req, *dataSet.UserID)
	if err != nil {
		if request.IsErrorUnauthorized(err) {
			dataServiceContext.RespondWithError(service.ErrorUnauthorized())
		} else {
			dataServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
		}
		return
	}
	if !permissions {
		dataServiceContext.RespondWithError(service.ErrorUnauthorized())
		return
	}

	if (dataSet.State != nil && *dataSet.State == "closed") || (dataSet.DataState != nil && *dataSet.DataState == "closed") { // TODO: Deprecated DataState (after data migration)
		dataServiceContext.RespondWithError(ErrorDataSetClosed(dataSetID))
		return
	}

	update := data.NewDataSetUpdate()
	if dataSet.DataSetType != nil && *dataSet.DataSetType == upload.DataSetTypeContinuous {
		details := request.DetailsFromContext(ctx)
		if !details.IsService() {
			dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			return
		}
		if err = request.DecodeRequestBody(req.Request, update); err != nil {
			request.MustNewResponder(res, req).Error(http.StatusBadRequest, err)
			return
		}
	} else {
		update.State = pointer.FromString(data.DataSetStateClosed)
	}

	dataSet, err = dataServiceContext.DataRepository().UpdateDataSet(ctx, dataSetID, update)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to update data set", err)
		return
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, dataSet)
}
