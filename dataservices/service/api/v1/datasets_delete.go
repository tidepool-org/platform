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

func DatasetsDelete(serviceContext service.Context) {
	datasetID := serviceContext.Request().PathParam("datasetid")
	if datasetID == "" {
		serviceContext.RespondWithError(ErrorDatasetIDMissing())
		return
	}

	dataset, err := serviceContext.DataStoreSession().GetDataset(datasetID)
	if err != nil {
		serviceContext.RespondWithError(ErrorDatasetIDNotFound(datasetID))
		return
	}

	targetUserID := dataset.UserID
	if targetUserID == "" {
		serviceContext.RespondWithInternalServerFailure("Unable to get user id from dataset")
		return
	}

	if !serviceContext.AuthenticationDetails().IsServer() {
		authenticatedUserID := serviceContext.AuthenticationDetails().UserID()

		var permissions client.Permissions
		permissions, err = serviceContext.UserServicesClient().GetUserPermissions(serviceContext, authenticatedUserID, targetUserID)
		if err != nil {
			if client.IsUnauthorizedError(err) {
				serviceContext.RespondWithError(commonService.ErrorUnauthorized())
			} else {
				serviceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[client.OwnerPermission]; !ok {
			if _, ok = permissions[client.CustodianPermission]; !ok {
				if _, ok = permissions[client.UploadPermission]; !ok || authenticatedUserID != dataset.ByUser {
					serviceContext.RespondWithError(commonService.ErrorUnauthorized())
					return
				}
			}
		}
	}

	if err = serviceContext.DataStoreSession().DeleteDataset(dataset); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to delete dataset", err)
		return
	}

	if err = serviceContext.MetricServicesClient().RecordMetric(serviceContext, "datasets_delete"); err != nil {
		serviceContext.Logger().WithError(err).Error("Unable to record metric")
	}

	serviceContext.RespondWithStatusAndData(http.StatusOK, struct{}{})
}
