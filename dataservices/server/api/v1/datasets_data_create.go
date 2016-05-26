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
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/deduplicator/root"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/dataservices/server"
	"github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/userservices/client"
)

func DatasetsDataCreate(serverContext server.Context) {
	datasetID := serverContext.Request().PathParam("datasetid")
	if datasetID == "" {
		serverContext.RespondWithError(ErrorDatasetIDMissing())
		return
	}

	// TODO: Improve serverContext.Store() Find - more specific
	var datasetUpload upload.Upload
	if err := serverContext.Store().Find(store.Query{"type": "upload", "uploadId": datasetID}, &datasetUpload); err != nil {
		serverContext.RespondWithError(ErrorDatasetIDNotFound(datasetID))
		return
	}

	// TODO: Validate
	targetUserID := datasetUpload.UserID
	targetGroupID := datasetUpload.GroupID

	err := serverContext.Client().ValidateTargetUserPermissions(serverContext, serverContext.RequestUserID(), targetUserID, client.UploadPermissions)
	if err != nil {
		if client.IsUnauthorizedError(err) {
			serverContext.RespondWithError(ErrorUnauthorized())
		} else {
			serverContext.RespondWithInternalServerFailure("Unable to validate target user permissions", err)
		}
		return
	}

	if datasetUpload.DataState != "open" {
		serverContext.RespondWithError(ErrorDatasetClosed(datasetID))
		return
	}

	deduplicator, err := root.NewFactory().NewDeduplicator(serverContext.Logger(), serverContext.Store(), &datasetUpload)
	if err != nil {
		serverContext.RespondWithInternalServerFailure("No duplicator found matching dataset", err)
		return
	}

	var rawDatumArray []interface{}
	if err = serverContext.Request().DecodeJsonPayload(&rawDatumArray); err != nil {
		serverContext.RespondWithError(ErrorJSONMalformed())
		return
	}

	datumArrayContext := context.NewStandard()

	datumArrayParser, err := parser.NewStandardArray(datumArrayContext, &rawDatumArray)
	if err != nil {
		serverContext.RespondWithInternalServerFailure("Unable to create datum parser", err)
		return
	}

	datumValidator, err := validator.NewStandard(datumArrayContext)
	if err != nil {
		serverContext.RespondWithInternalServerFailure("Unable to create datum validator", err)
		return
	}

	datumNormalizer, err := normalizer.NewStandard(datumArrayContext)
	if err != nil {
		serverContext.RespondWithInternalServerFailure("Unable to create datum normalizer", err)
		return
	}

	datumArray := []data.Datum{}
	for index := range *datumArrayParser.Array() {
		var datum data.Datum
		datum, err = types.Parse(datumArrayContext.NewChildContext(index), datumArrayParser.NewChildObjectParser(index))
		if err != nil {
			serverContext.RespondWithInternalServerFailure("Unable to parse datum", err)
			return
		}

		datum.Validate(datumValidator.NewChildValidator(index))
		datumArray = append(datumArray, datum)
	}

	if errors := datumArrayContext.Errors(); len(errors) > 0 {
		serverContext.RespondWithStatusAndErrors(http.StatusBadRequest, errors)
		return
	}

	for _, datum := range datumArray {
		datum.SetUserID(targetUserID)
		datum.SetGroupID(targetGroupID)
		datum.SetDatasetID(datasetID)
		datum.Normalize(datumNormalizer)
	}

	datumArray = append(datumArray, datumNormalizer.Data()...)

	if err = deduplicator.AddDataToDataset(datumArray); err != nil {
		serverContext.RespondWithInternalServerFailure("Unable to add data to dataset", err)
		return
	}

	serverContext.Response().WriteHeader(http.StatusOK)
}
