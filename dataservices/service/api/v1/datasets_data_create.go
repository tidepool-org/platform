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
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/dataservices/service"
	"github.com/tidepool-org/platform/userservices/client"
)

func DatasetsDataCreate(serviceContext service.Context) {
	datasetID := serviceContext.Request().PathParam("datasetid")
	if datasetID == "" {
		serviceContext.RespondWithError(ErrorDatasetIDMissing())
		return
	}

	dataset, err := serviceContext.DataStoreSession().GetDataset(datasetID)
	if err != nil {
		serviceContext.RespondWithError(ErrorDatasetIDNotFound(datasetID))
		return
	}

	permissions, err := serviceContext.UserServicesClient().GetUserPermissions(serviceContext, serviceContext.RequestUserID(), dataset.UserID)
	if err != nil {
		if client.IsUnauthorizedError(err) {
			serviceContext.RespondWithError(ErrorUnauthorized())
		} else {
			serviceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
		}
		return
	}
	if _, ok := permissions[client.UploadPermission]; !ok {
		serviceContext.RespondWithError(ErrorUnauthorized())
		return
	}

	if dataset.DataState != "open" {
		serviceContext.RespondWithError(ErrorDatasetClosed(datasetID))
		return
	}

	deduplicator, err := serviceContext.DataDeduplicatorFactory().NewDeduplicator(serviceContext.Logger(), serviceContext.DataStoreSession(), dataset)
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("No duplicator found matching dataset", err)
		return
	}

	var rawDatumArray []interface{}
	if err = serviceContext.Request().DecodeJsonPayload(&rawDatumArray); err != nil {
		serviceContext.RespondWithError(ErrorJSONMalformed())
		return
	}

	datumArrayContext, err := context.NewStandard(serviceContext.Logger())
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to create datum array context", err)
		return
	}

	datumArrayParser, err := parser.NewStandardArray(datumArrayContext, serviceContext.DataFactory(), &rawDatumArray, parser.AppendErrorNotParsed)
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to create datum array parser", err)
		return
	}

	datumValidator, err := validator.NewStandard(datumArrayContext)
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to create datum validator", err)
		return
	}

	datumNormalizer, err := normalizer.NewStandard(datumArrayContext)
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to create datum normalizer", err)
		return
	}

	datumArray := []data.Datum{}
	for index := range *datumArrayParser.Array() {
		if datum := datumArrayParser.ParseDatum(index); datum != nil && *datum != nil {
			(*datum).Validate(datumValidator.NewChildValidator(index))
			datumArray = append(datumArray, *datum)
		}
	}

	datumArrayParser.ProcessNotParsed()

	if errors := datumArrayContext.Errors(); len(errors) > 0 {
		serviceContext.RespondWithStatusAndErrors(http.StatusBadRequest, errors)
		return
	}

	for _, datum := range datumArray {
		datum.Normalize(datumNormalizer)
	}

	datumArray = append(datumArray, datumNormalizer.Data()...)

	if err = deduplicator.AddDataToDataset(datumArray); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to add data to dataset", err)
		return
	}

	serviceContext.RespondWithStatusAndData(http.StatusOK, []struct{}{})
}
