package v1

import (
	"net/http"
	"strconv"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
	userClient "github.com/tidepool-org/platform/user/client"
)

func DatasetsDataCreate(dataServiceContext dataService.Context) {
	datasetID := dataServiceContext.Request().PathParam("datasetid")
	if datasetID == "" {
		dataServiceContext.RespondWithError(ErrorDatasetIDMissing())
		return
	}

	dataset, err := dataServiceContext.DataSession().GetDatasetByID(datasetID)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get dataset by id", err)
		return
	}
	if dataset == nil {
		dataServiceContext.RespondWithError(ErrorDatasetIDNotFound(datasetID))
		return
	}

	if !dataServiceContext.AuthDetails().IsServer() {
		var permissions userClient.Permissions
		permissions, err = dataServiceContext.UserClient().GetUserPermissions(dataServiceContext, dataServiceContext.AuthDetails().UserID(), dataset.UserID)
		if err != nil {
			if client.IsUnauthorizedError(err) {
				dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			} else {
				dataServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[userClient.UploadPermission]; !ok {
			dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			return
		}
	}

	if dataset.State == "closed" || dataset.DataState == "closed" { // TODO: Deprecated DataState (after data migration)
		dataServiceContext.RespondWithError(ErrorDatasetClosed(datasetID))
		return
	}

	deduplicator, err := dataServiceContext.DataDeduplicatorFactory().NewRegisteredDeduplicatorForDataset(dataServiceContext.Logger(), dataServiceContext.DataSession(), dataset)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create registered deduplicator for dataset", err)
		return
	}

	var rawDatumArray []interface{}
	if err = dataServiceContext.Request().DecodeJsonPayload(&rawDatumArray); err != nil {
		dataServiceContext.RespondWithError(service.ErrorJSONMalformed())
		return
	}

	datumArrayContext, err := context.NewStandard(dataServiceContext.Logger())
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create datum array context", err)
		return
	}

	datumArrayParser, err := parser.NewStandardArray(datumArrayContext, dataServiceContext.DataFactory(), &rawDatumArray, parser.AppendErrorNotParsed)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create datum array parser", err)
		return
	}

	datumValidator, err := validator.NewStandard(datumArrayContext)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create datum validator", err)
		return
	}

	datumNormalizer, err := normalizer.NewStandard(datumArrayContext)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create datum normalizer", err)
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
		dataServiceContext.RespondWithStatusAndErrors(http.StatusBadRequest, errors)
		return
	}

	for _, datum := range datumArray {
		datum.Normalize(datumNormalizer)
	}

	datumArray = append(datumArray, datumNormalizer.Data()...)

	for _, datum := range datumArray {
		datum.SetUserID(dataset.UserID)
		datum.SetDatasetID(dataset.UploadID)
	}

	if err = deduplicator.AddDatasetData(datumArray); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to add dataset data", err)
		return
	}

	if err = dataServiceContext.MetricClient().RecordMetric(dataServiceContext, "datasets_data_create", map[string]string{"count": strconv.Itoa(len(datumArray))}); err != nil {
		dataServiceContext.Logger().WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, []struct{}{})
}
