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

func DatasetsUpdate(serviceContext service.Context) {
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

	permissions, err := serviceContext.UserServicesClient().GetUserPermissions(serviceContext, serviceContext.RequestUserID(), dataset.UserID)
	if err != nil {
		if client.IsUnauthorizedError(err) {
			serviceContext.RespondWithError(ErrorUnauthorized())
		} else {
			serviceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
		}
		return
	}
	if _, ok := permissions[client.UploadPermission]; !ok {
		serviceContext.RespondWithError(ErrorUnauthorized())
		return
	}

	if dataset.DataState != "open" {
		serviceContext.RespondWithError(ErrorDatasetClosed(datasetID))
		return
	}

	dataset.SetDataState("closed")

	if err = serviceContext.DataStoreSession().UpdateDataset(dataset); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to update dataset", err)
		return
	}

	deduplicator, err := serviceContext.DataDeduplicatorFactory().NewDeduplicator(serviceContext.Logger(), serviceContext.DataStoreSession(), dataset)
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("No duplicator found matching dataset", err)
		return
	}

	if err = deduplicator.FinalizeDataset(); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to finalize dataset", err)
		return
	}

	serviceContext.Response().WriteHeader(http.StatusOK)
	serviceContext.Response().WriteJson(dataset)
}
