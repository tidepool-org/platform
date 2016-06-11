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

	"github.com/tidepool-org/platform/data/deduplicator/root"
	"github.com/tidepool-org/platform/dataservices/server"
	"github.com/tidepool-org/platform/userservices/client"
)

func DatasetsUpdate(serverContext server.Context) {
	datasetID := serverContext.Request().PathParam("datasetid")
	if datasetID == "" {
		serverContext.RespondWithError(ErrorDatasetIDMissing())
		return
	}

	dataset, err := serverContext.DataStoreSession().GetDataset(datasetID)
	if err != nil {
		serverContext.RespondWithError(ErrorDatasetIDNotFound(datasetID))
		return
	}

	err = serverContext.UserServicesClient().ValidateTargetUserPermissions(serverContext, serverContext.RequestUserID(), dataset.UserID, client.UploadPermissions)
	if err != nil {
		if client.IsUnauthorizedError(err) {
			serverContext.RespondWithError(ErrorUnauthorized())
		} else {
			serverContext.RespondWithInternalServerFailure("Unable to validate target user permissions", err)
		}
		return
	}

	if dataset.DataState != "open" {
		serverContext.RespondWithError(ErrorDatasetClosed(datasetID))
		return
	}

	dataset.SetDataState("closed")

	if err = serverContext.DataStoreSession().UpdateDataset(dataset); err != nil {
		serverContext.RespondWithInternalServerFailure("Unable to update dataset", err)
		return
	}

	deduplicator, err := root.NewFactory().NewDeduplicator(serverContext.Logger(), serverContext.DataStoreSession(), dataset)
	if err != nil {
		serverContext.RespondWithInternalServerFailure("No duplicator found matching dataset", err)
		return
	}

	if err = deduplicator.FinalizeDataset(); err != nil {
		serverContext.RespondWithInternalServerFailure("Unable to finalize dataset", err)
		return
	}

	serverContext.Response().WriteHeader(http.StatusOK)
	serverContext.Response().WriteJson(dataset)
}
