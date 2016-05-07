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
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/dataservices/server/api"
	"github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/userservices/client"
)

func DatasetsUpdate(context *api.Context) {
	datasetID := context.Request().PathParam("datasetid")
	if datasetID == "" {
		context.RespondWithError(ErrorDatasetIDMissing())
		return
	}

	// TODO: Improve context.Store() Find - more specific
	var datasetUpload upload.Upload
	if err := context.Store().Find(store.Query{"type": "upload", "uploadId": datasetID}, &datasetUpload); err != nil {
		context.RespondWithError(ErrorDatasetIDNotFound(datasetID))
		return
	}

	// TODO: Validate
	targetUserID := *datasetUpload.UserID

	err := context.Client().ValidateTargetUserPermissions(context.Context, context.RequestUserID, targetUserID, client.UploadPermissions)
	if err != nil {
		if client.IsUnauthorizedError(err) {
			context.RespondWithError(ErrorUnauthorized())
		} else {
			context.RespondWithServerFailure("Unable to validate target user permissions", err)
		}
		return
	}

	if datasetUpload.DataState == nil || *datasetUpload.DataState != "open" {
		context.RespondWithError(ErrorDatasetClosed(datasetID))
		return
	}

	dataState := "closed"
	datasetUpload.DataState = &dataState

	if err := context.Store().Update(map[string]interface{}{"type": "upload", "uploadId": datasetID}, datasetUpload); err != nil {
		context.RespondWithServerFailure("Unable to insert dataset", err)
		return
	}

	// TODO: Pass in logger here
	deduplicator, err := root.NewFactory().NewDeduplicator(&datasetUpload, context.Store(), context.Logger())
	if err != nil {
		context.RespondWithServerFailure("No duplicator found matching dataset", err)
		return
	}

	if err := deduplicator.FinalizeDataset(); err != nil {
		context.RespondWithServerFailure("Unable to finalize dataset", err)
		return
	}

	// TODO: Filter datasetUpload to only "public" fields
	context.Response().WriteHeader(http.StatusOK)
	context.Response().WriteJson(datasetUpload)
}
