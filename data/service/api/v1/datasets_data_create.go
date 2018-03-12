package v1

import (
	"net/http"
	"strconv"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	dataService "github.com/tidepool-org/platform/data/service"
	dataTypesFactory "github.com/tidepool-org/platform/data/types/factory"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

func DatasetsDataCreate(dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	lgr := log.LoggerFromContext(ctx)

	datasetID := dataServiceContext.Request().PathParam("dataSetId")
	if datasetID == "" {
		dataServiceContext.RespondWithError(ErrorDatasetIDMissing())
		return
	}

	dataset, err := dataServiceContext.DataSession().GetDatasetByID(ctx, datasetID)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get dataset by id", err)
		return
	}
	if dataset == nil {
		dataServiceContext.RespondWithError(ErrorDatasetIDNotFound(datasetID))
		return
	}

	if details := request.DetailsFromContext(ctx); !details.IsService() {
		var permissions user.Permissions
		permissions, err = dataServiceContext.UserClient().GetUserPermissions(ctx, details.UserID(), *dataset.UserID)
		if err != nil {
			if errors.Code(err) == request.ErrorCodeUnauthorized {
				dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			} else {
				dataServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[user.UploadPermission]; !ok {
			dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			return
		}
	}

	if (dataset.State != nil && *dataset.State == "closed") || (dataset.DataState != nil && *dataset.DataState == "closed") { // TODO: Deprecated DataState (after data migration)
		dataServiceContext.RespondWithError(ErrorDatasetClosed(datasetID))
		return
	}

	deduplicator, err := dataServiceContext.DataDeduplicatorFactory().NewRegisteredDeduplicatorForDataset(lgr, dataServiceContext.DataSession(), dataset)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create registered deduplicator for dataset", err)
		return
	}

	var rawDatumArray []interface{}
	if err = dataServiceContext.Request().DecodeJsonPayload(&rawDatumArray); err != nil {
		dataServiceContext.RespondWithError(service.ErrorJSONMalformed())
		return
	}

	datumArrayContext, err := context.NewStandard(lgr)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create datum array context", err)
		return
	}

	datumArrayParser, err := parser.NewStandardArray(datumArrayContext, &rawDatumArray, parser.AppendErrorNotParsed)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create datum array parser", err)
		return
	}

	validator := structureValidator.New()
	normalizer := dataNormalizer.New()

	datumArray := []data.Datum{}
	for index := range *datumArrayParser.Array() {
		reference := strconv.Itoa(index)
		if datum := dataTypesFactory.ParseDatum(datumArrayParser.NewChildObjectParser(index)); datum != nil && *datum != nil {
			(*datum).Validate(validator.WithReference(reference))
			(*datum).Normalize(normalizer.WithReference(reference))
			datumArray = append(datumArray, *datum)
		}
	}

	datumArrayParser.ProcessNotParsed()

	if errs := datumArrayContext.Errors(); len(errs) > 0 {
		dataServiceContext.RespondWithStatusAndErrors(http.StatusBadRequest, errs)
		return
	}

	if err = validator.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
	}
	if err = normalizer.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
	}

	datumArray = append(datumArray, normalizer.Data()...)

	for _, datum := range datumArray {
		datum.SetUserID(dataset.UserID)
		datum.SetDatasetID(dataset.UploadID)
	}

	if err = deduplicator.AddDatasetData(ctx, datumArray); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to add dataset data", err)
		return
	}

	if err = dataServiceContext.MetricClient().RecordMetric(ctx, "datasets_data_create", map[string]string{"count": strconv.Itoa(len(datumArray))}); err != nil {
		lgr.WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, []struct{}{})
}
