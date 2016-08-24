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
		serviceContext.RespondWithInternalServerFailure("Unable to get user id from dataset", err)
		return
	}

	permissions, err := serviceContext.UserServicesClient().GetUserPermissions(serviceContext, serviceContext.RequestUserID(), targetUserID)
	if err != nil {
		if client.IsUnauthorizedError(err) {
			serviceContext.RespondWithError(ErrorUnauthorized())
		} else {
			serviceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
		}
		return
	}
	if _, ok := permissions[client.OwnerPermission]; !ok {
		if _, ok = permissions[client.CustodianPermission]; !ok {
			if _, ok = permissions[client.UploadPermission]; !ok || serviceContext.RequestUserID() != targetUserID {
				serviceContext.RespondWithError(ErrorUnauthorized())
				return
			}
		}
	}

	if err = serviceContext.DataStoreSession().DeleteDataset(datasetID); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to delete dataset", err)
		return
	}

	serviceContext.RespondWithStatusAndData(http.StatusOK, struct{}{})
}
