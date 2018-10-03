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
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

func DataSetsDataCreate(dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	lgr := log.LoggerFromContext(ctx)

	dataSetID := dataServiceContext.Request().PathParam("dataSetId")
	if dataSetID == "" {
		dataServiceContext.RespondWithError(ErrorDataSetIDMissing())
		return
	}

	dataSet, err := dataServiceContext.DataSession().GetDataSetByID(ctx, dataSetID)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get data set by id", err)
		return
	}
	if dataSet == nil {
		dataServiceContext.RespondWithError(ErrorDataSetIDNotFound(dataSetID))
		return
	}

	if details := request.DetailsFromContext(ctx); !details.IsService() {
		var permissions user.Permissions
		permissions, err = dataServiceContext.UserClient().GetUserPermissions(ctx, details.UserID(), *dataSet.UserID)
		if err != nil {
			if request.IsErrorUnauthorized(err) {
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

	if (dataSet.State != nil && *dataSet.State == "closed") || (dataSet.DataState != nil && *dataSet.DataState == "closed") { // TODO: Deprecated DataState (after data migration)
		dataServiceContext.RespondWithError(ErrorDataSetClosed(dataSetID))
		return
	}

	deduplicator, err := dataServiceContext.DataDeduplicatorFactory().NewRegisteredDeduplicatorForDataSet(lgr, dataServiceContext.DataSession(), dataSet)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create registered deduplicator for data set", err)
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
		return
	}
	if err = normalizer.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
		return
	}

	datumArray = append(datumArray, normalizer.Data()...)

	for _, datum := range datumArray {
		datum.SetUserID(dataSet.UserID)
		datum.SetDataSetID(dataSet.UploadID)
	}

	if err = deduplicator.AddDataSetData(ctx, datumArray); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to add data set data", err)
		return
	}

	if err = dataServiceContext.MetricClient().RecordMetric(ctx, "data_sets_data_create", map[string]string{"count": strconv.Itoa(len(datumArray))}); err != nil {
		lgr.WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, []struct{}{})
}
