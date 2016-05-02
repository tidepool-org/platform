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

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator/root"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/store"
)

// TODO: Move this to more common location
const (
	ParamDatasetID = "datasetid"
	ParamUserID    = "userid"
)

func (s *Server) DatasetDataCreate(context *Context) {
	// TODO: Further validation of datasetID
	datasetID := context.Request().PathParam(ParamDatasetID)
	if datasetID == "" {
		context.RespondWithError(ConstructError(ErrorDatasetIDMalformed, datasetID))
		return
	}

	// TODO: Improve context.Store() Find - more specific
	var datasetUpload upload.Upload
	if err := context.Store().Find(store.Query{"type": "upload", "uploadId": datasetID}, &datasetUpload); err != nil {
		context.RespondWithError(ConstructError(ErrorDatasetIDNotFound, datasetID))
		return
	}

	userID := *datasetUpload.UserID
	groupID := datasetUpload.GroupID

	// if !checkPermisson(context.Request(), user.Permission{}) {
	// 	rest.Error(context.Response, missingPermissionsError, http.StatusUnauthorized)
	// 	return
	// }

	if datasetUpload.DataState == nil || *datasetUpload.DataState != "open" {
		context.RespondWithError(ConstructError(ErrorDatasetClosed, datasetID))
		return
	}

	// TODO: Pass in logger here
	deduplicator, err := root.NewFactory().NewDeduplicator(&datasetUpload, context.Store(), context.Logger())
	if err != nil {
		context.RespondWithServerFailure("No duplicator found matching dataset", err)
		return
	}

	var rawDatumArray types.DatumArray
	if err := context.Request().DecodeJsonPayload(&rawDatumArray); err != nil {
		context.RespondWithError(ConstructError(ErrorJSONMalformed))
		return
	}

	// TODO: Fix common data
	commonData := map[string]interface{}{types.BaseUserIDField.Name: userID, types.BaseGroupIDField.Name: groupID, "uploadId": datasetID, "_active": false, "_schemaVersion": 1}
	datumArray, errors := data.NewTypeBuilder(commonData).BuildFromDatumArray(rawDatumArray)
	if errors != nil {
		context.RespondWithStatusAndErrors(http.StatusBadRequest, errors)
		return
	}

	if err := deduplicator.AddDataToDataset(datumArray); err != nil {
		context.RespondWithServerFailure("Unable to add data to dataset", err)
		return
	}

	context.Response().WriteHeader(http.StatusOK)
}
