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
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/dataservices/server"
	"github.com/tidepool-org/platform/userservices/client"
)

func DatasetsDataCreate(serverContext server.Context) {
	datasetID := serverContext.Request().PathParam("datasetid")
	if datasetID == "" {
		serverContext.RespondWithError(ErrorDatasetIDMissing())
		return
	}

	datasetUpload, err := serverContext.DataStoreSession().GetDataset(datasetID)
	if err != nil {
		serverContext.RespondWithError(ErrorDatasetIDNotFound(datasetID))
		return
	}

	// TODO: Validate
	targetUserID := datasetUpload.UserID
	targetGroupID := datasetUpload.GroupID

	err = serverContext.UserServicesClient().ValidateTargetUserPermissions(serverContext, serverContext.RequestUserID(), targetUserID, client.UploadPermissions)
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

	deduplicator, err := root.NewFactory().NewDeduplicator(serverContext.Logger(), serverContext.DataStoreSession(), datasetUpload)
	if err != nil {
		serverContext.RespondWithInternalServerFailure("No duplicator found matching dataset", err)
		return
	}

	var rawDatumArray []interface{}
	if err = serverContext.Request().DecodeJsonPayload(&rawDatumArray); err != nil {
		serverContext.RespondWithError(ErrorJSONMalformed())
		return
	}

	datumArrayContext, err := context.NewStandard(serverContext.Logger())
	if err != nil {
		serverContext.RespondWithInternalServerFailure("Unable to create datum array context", err)
		return
	}

	datumArrayParser, err := parser.NewStandardArray(datumArrayContext, &rawDatumArray, parser.AppendErrorNotParsed)
	if err != nil {
		serverContext.RespondWithInternalServerFailure("Unable to create datum array parser", err)
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
		datumObjectParser := datumArrayParser.NewChildObjectParser(index)
		datum, datumErr := types.Parse(datumObjectParser)
		if datumErr != nil {
			serverContext.RespondWithInternalServerFailure("Unable to parse datum", datumErr)
			return
		}
		datumObjectParser.ProcessNotParsed()

		if datum != nil {
			datum.Validate(datumValidator.NewChildValidator(index))
			datumArray = append(datumArray, datum)
		}
	}

	datumArrayParser.ProcessNotParsed()

	if errors := datumArrayContext.Errors(); len(errors) > 0 {
		serverContext.RespondWithStatusAndErrors(http.StatusBadRequest, errors)
		return
	}

	for _, datum := range datumArray {
		datum.Normalize(datumNormalizer)
	}

	datumArray = append(datumArray, datumNormalizer.Data()...)

	for _, datum := range datumArray {
		datum.SetUserID(targetUserID)
		datum.SetGroupID(targetGroupID)
		datum.SetDatasetID(datasetID)
	}

	if err = deduplicator.AddDataToDataset(datumArray); err != nil {
		serverContext.RespondWithInternalServerFailure("Unable to add data to dataset", err)
		return
	}

	serverContext.Response().WriteHeader(http.StatusOK)
}
