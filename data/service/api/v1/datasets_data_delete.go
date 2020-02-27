package v1

import (
	"net/http"

	"github.com/tidepool-org/platform/data"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
)

// DataSetsDataDelete godoc
// @Summary Delete user data
// @Description Caller must be a service, the owner, or have the authorizations to do it in behalf of the user.
// @ID platform-data-api-DataSetsDataDelete
// @Accept json
// @Param dataSetId path string true "dataSet ID"
// @Security TidepoolSessionToken
// @Security TidepoolServiceSecret
// @Success 200 "Operation is a success"
// @Failure 400 {object} service.Error "dataSet ID is missing"
// @Failure 403 {object} service.Error "Forbiden: caller is not a service"
// @Failure 409 {object} service.Error "Data set with specified id is closed"
// @Failure 500 {object} service.Error "Unable to perform the operation"
// @Router /v1/data_sets/:dataSetId/data [delete]
func DataSetsDataDelete(dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()

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

	if details := request.DetailsFromContext(ctx); !details.IsService() {
		var permissions permission.Permissions
		permissions, err = dataServiceContext.PermissionClient().GetUserPermissions(ctx, details.UserID(), *dataSet.UserID)
		if err != nil {
			if request.IsErrorUnauthorized(err) {
				dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			} else {
				dataServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[permission.Write]; !ok {
			dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			return
		}
	}

	if (dataSet.State != nil && *dataSet.State == "closed") || (dataSet.DataState != nil && *dataSet.DataState == "closed") { // TODO: Deprecated DataState (after data migration)
		dataServiceContext.RespondWithError(ErrorDataSetClosed(dataSetID))
		return
	}

	selectors := data.NewSelectors()
	if err = request.DecodeRequestBody(dataServiceContext.Request().Request, selectors); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("unable to parse selectors", err)
		return
	}

	if deduplicator, getErr := dataServiceContext.DataDeduplicatorFactory().Get(dataSet); getErr != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get deduplicator", getErr)
		return
	} else if deduplicator == nil {
		dataServiceContext.RespondWithInternalServerFailure("Deduplicator not found")
		return
	} else if err = deduplicator.DeleteData(ctx, dataServiceContext.DataSession(), dataSet, selectors); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to delete data", err)
		return
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, []struct{}{})
}
