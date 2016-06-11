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

func DatasetsUpdate(context server.Context) {
	datasetID := context.Request().PathParam("datasetid")
	if datasetID == "" {
		context.RespondWithError(ErrorDatasetIDMissing())
		return
	}

	dataset, err := context.DataStoreSession().GetDataset(datasetID)
	if err != nil {
		context.RespondWithError(ErrorDatasetIDNotFound(datasetID))
		return
	}

	// TODO: Validate
	targetUserID := dataset.UserID

	err = context.UserServicesClient().ValidateTargetUserPermissions(context, context.RequestUserID(), targetUserID, client.UploadPermissions)
	if err != nil {
		if client.IsUnauthorizedError(err) {
			context.RespondWithError(ErrorUnauthorized())
		} else {
			context.RespondWithInternalServerFailure("Unable to validate target user permissions", err)
		}
		return
	}

	if dataset.DataState != "open" {
		context.RespondWithError(ErrorDatasetClosed(datasetID))
		return
	}

	dataset.SetDataState("closed")

	if err = context.DataStoreSession().UpdateDataset(dataset); err != nil {
		context.RespondWithInternalServerFailure("Unable to update dataset", err)
		return
	}

	deduplicator, err := root.NewFactory().NewDeduplicator(context.Logger(), context.DataStoreSession(), dataset)
	if err != nil {
		context.RespondWithInternalServerFailure("No duplicator found matching dataset", err)
		return
	}

	if err = deduplicator.FinalizeDataset(); err != nil {
		context.RespondWithInternalServerFailure("Unable to finalize dataset", err)
		return
	}

	// TODO: Filter dataset to only "public" fields
	context.Response().WriteHeader(http.StatusOK)
	context.Response().WriteJson(dataset)
}
