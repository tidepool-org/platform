package dataservices

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

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator/root"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/user"
)

func (s *Server) DatasetCreate(context *Context) {
	// TODO: Further validation of userID
	userID := context.Request().PathParam(ParamUserID)
	if userID == "" {
		context.RespondWithError(ConstructError(ErrorUserIDMalformed, userID))
		return
	}

	// if !checkPermisson(context.Request(), user.Permission{}) {
	// 	rest.Error(context.Response(), missingPermissionsError, http.StatusUnauthorized) // TODO: JSON error
	// 	return
	// }

	groupID := context.Request().Env[user.GROUPID]
	if groupID == "" {
		rest.Error(context.Response(), "Group id is missing", http.StatusBadRequest) // TODO: Fix this
		return
	}

	// TODO: Do we need to do this? Shouldn't we fail on no group ID earlier and let context.Request().ContentLength fail when decoding JSON?
	// if context.Request().ContentLength == 0 || groupID == "" {
	// 	rest.Error(context.Response(), missingDataError, http.StatusBadRequest) // TODO: JSON error
	// 	return
	// }

	var rawDatasetDatum types.Datum
	if err := context.Request().DecodeJsonPayload(&rawDatasetDatum); err != nil {
		context.RespondWithError(ConstructError(ErrorJSONMalformed))
		return
	}

	// TODO: Not sure about how best to represent these constants?
	// TODO: Move uploadId and dataState into type builder (verify not there originally)
	commonDatum := types.Datum{types.BaseUserIDField.Name: userID, types.BaseGroupIDField.Name: groupID}
	datasetBuiltDatum, errors := data.NewTypeBuilder(commonDatum).BuildFromDatum(rawDatasetDatum)
	if errors != nil {
		context.RespondWithStatusAndErrors(http.StatusBadRequest, errors)
		return
	}

	datasetUpload, ok := datasetBuiltDatum.(*upload.Upload)
	if !ok {
		context.RespondWithServerFailure("Unexpected datum type", datasetBuiltDatum)
		return
	}

	// TODO: Move this to a better location
	uploadID := app.NewUUID()
	dataState := "open"

	datasetUpload.UploadID = &uploadID
	datasetUpload.DataState = &dataState

	if err := context.Store().Insert(datasetUpload); err != nil {
		context.RespondWithServerFailure("Unable to insert dataset", err)
		return
	}

	// TODO: Pass in logger here
	deduplicator, err := root.NewFactory().NewDeduplicator(datasetUpload, context.Store(), context.Logger())
	if err != nil {
		context.RespondWithServerFailure("No duplicator found matching dataset", err)
		return
	}

	if err := deduplicator.InitializeDataset(); err != nil {
		context.RespondWithServerFailure("Unable to initialize dataset", err)
		return
	}

	// TODO: Filter datasetUpload to only "public" fields
	context.Response().WriteHeader(http.StatusCreated)
	context.Response().WriteJson(datasetUpload)
}
