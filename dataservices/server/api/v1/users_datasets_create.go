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

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator/root"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/dataservices/server"
	"github.com/tidepool-org/platform/userservices/client"
)

func UsersDatasetsCreate(context server.Context) {
	targetUserID := context.Request().PathParam("userid")
	if targetUserID == "" {
		context.RespondWithError(ErrorUserIDMissing())
		return
	}

	err := context.Client().ValidateTargetUserPermissions(context, context.RequestUserID(), targetUserID, client.UploadPermissions)
	if err != nil {
		if client.IsUnauthorizedError(err) {
			context.RespondWithError(ErrorUnauthorized())
		} else {
			context.RespondWithInternalServerFailure("Unable to validate target user permissions", err)
		}
		return
	}

	targetUserGroupID, err := context.Client().GetUserGroupID(context, targetUserID)
	if err != nil {
		if client.IsUnauthorizedError(err) {
			context.RespondWithError(ErrorUnauthorized())
		} else {
			context.RespondWithInternalServerFailure("Unable to get group id for target user", err)
		}
		return
	}

	var rawDatasetDatum types.Datum
	if err = context.Request().DecodeJsonPayload(&rawDatasetDatum); err != nil {
		context.RespondWithError(ErrorJSONMalformed())
		return
	}

	// TODO: Not sure about how best to represent these constants?
	// TODO: Move uploadId and dataState into type builder (verify not there originally)
	commonDatum := types.Datum{types.BaseUserIDField.Name: targetUserID, types.BaseGroupIDField.Name: targetUserGroupID}
	datasetBuiltDatum, errors := data.NewTypeBuilder(commonDatum).BuildFromDatum(rawDatasetDatum)
	if errors != nil {
		context.RespondWithStatusAndErrors(http.StatusBadRequest, errors)
		return
	}

	datasetUpload, ok := datasetBuiltDatum.(*upload.Upload)
	if !ok {
		context.RespondWithInternalServerFailure("Unexpected datum type", datasetBuiltDatum)
		return
	}

	// TODO: Move this to a better location
	uploadID := app.NewUUID()
	dataState := "open"

	datasetUpload.UploadID = &uploadID
	datasetUpload.DataState = &dataState

	if err = context.Store().Insert(datasetUpload); err != nil {
		context.RespondWithInternalServerFailure("Unable to insert dataset", err)
		return
	}

	// TODO: Pass in logger here
	deduplicator, err := root.NewFactory().NewDeduplicator(context.Logger(), context.Store(), datasetUpload)
	if err != nil {
		context.RespondWithInternalServerFailure("No duplicator found matching dataset", err)
		return
	}

	if err = deduplicator.InitializeDataset(); err != nil {
		context.RespondWithInternalServerFailure("Unable to initialize dataset", err)
		return
	}

	// TODO: Filter datasetUpload to only "public" fields
	context.Response().WriteHeader(http.StatusCreated)
	context.Response().WriteJson(datasetUpload)
}
