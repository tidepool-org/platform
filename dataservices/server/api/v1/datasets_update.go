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
	"github.com/tidepool-org/platform/dataservices/server/api/v1/errors"
	"github.com/tidepool-org/platform/store"
)

func DatasetsUpdate(context *api.Context) {
	// TODO: Further validation of datasetID
	datasetID := context.Request().PathParam(ParamDatasetID)
	if datasetID == "" {
		context.RespondWithError(errors.ConstructError(errors.DatasetIDMissing))
		return
	}

	// TODO: Improve context.Store() Find - more specific
	var datasetUpload upload.Upload
	if err := context.Store().Find(store.Query{"type": "upload", "uploadId": datasetID}, &datasetUpload); err != nil {
		context.RespondWithError(errors.ConstructError(errors.DatasetIDNotFound, datasetID))
		return
	}

	if datasetUpload.DataState == nil || *datasetUpload.DataState != "open" {
		context.RespondWithError(errors.ConstructError(errors.DatasetClosed, datasetID))
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
