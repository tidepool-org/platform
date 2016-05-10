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

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator/root"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/dataservices/server/api"
	"github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/userservices/client"
)

func DatasetsDataCreate(context *api.Context) {
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
	targetGroupID := datasetUpload.GroupID

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

	// TODO: Pass in logger here
	deduplicator, err := root.NewFactory().NewDeduplicator(&datasetUpload, context.Store(), context.Logger())
	if err != nil {
		context.RespondWithServerFailure("No duplicator found matching dataset", err)
		return
	}

	var rawDatumArray types.DatumArray
	if err = context.Request().DecodeJsonPayload(&rawDatumArray); err != nil {
		context.RespondWithError(ErrorJSONMalformed())
		return
	}

	// TODO: Fix common data
	commonData := map[string]interface{}{types.BaseUserIDField.Name: targetUserID, types.BaseGroupIDField.Name: targetGroupID, "uploadId": datasetID, "_active": false, "_schemaVersion": 1}
	datumArray, errors := data.NewTypeBuilder(commonData).BuildFromDatumArray(rawDatumArray)
	if errors != nil {
		context.RespondWithStatusAndErrors(http.StatusBadRequest, errors)
		return
	}

	if err = deduplicator.AddDataToDataset(datumArray); err != nil {
		context.RespondWithServerFailure("Unable to add data to dataset", err)
		return
	}

	context.Response().WriteHeader(http.StatusOK)
}
