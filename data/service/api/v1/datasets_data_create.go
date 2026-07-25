package v1

import (
	"net/http"
	"strconv"

	"github.com/tidepool-org/platform/data"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataService "github.com/tidepool-org/platform/data/service"
	dataTypesFactory "github.com/tidepool-org/platform/data/types/factory"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/summary"
	summaryTypes "github.com/tidepool-org/platform/summary/types"
)

func DataSetsDataCreate(dataServiceContext dataService.Context) {
	req := dataServiceContext.Request()
	ctx := req.Context()
	lgr := log.LoggerFromContext(ctx)

	dataSetID := dataServiceContext.Request().PathParam("dataSetId")
	if dataSetID == "" {
		dataServiceContext.RespondWithError(ErrorDataSetIDMissing())
		return
	}

	dataSet, err := dataServiceContext.DataRepository().GetDataSet(ctx, dataSetID)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get data set by id", err)
		return
	} else if dataSet == nil {
		dataServiceContext.RespondWithError(ErrorDataSetIDNotFound(dataSetID))
		return
	}

	var authDetails request.AuthDetails
	if authDetails = request.GetAuthDetails(ctx); !authDetails.IsService() {
		var permissions permission.Permissions
		permissions, err = dataServiceContext.PermissionClient().GetUserPermissions(ctx, authDetails.UserID(), *dataSet.UserID)
		if err != nil {
			if request.IsErrorUnauthorized(err) {
				dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			} else {
				dataServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[permission.Write]; !ok {
			dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			return
		}
	}

	if (dataSet.State != nil && *dataSet.State == "closed") || (dataSet.DataState != nil && *dataSet.DataState == "closed") { // TODO: Deprecated DataState (after data migration)
		dataServiceContext.RespondWithError(ErrorDataSetClosed(dataSetID))
		return
	}

	var rawDatumArray []any
	if err = dataServiceContext.Request().DecodeJsonPayload(&rawDatumArray); err != nil {
		dataServiceContext.RespondWithError(service.ErrorJSONMalformed())
		return
	}

	logger := log.LoggerFromContext(ctx)
	parser := structureParser.NewArray(logger, &rawDatumArray)
	validator := structureValidator.New(logger)
	normalizer := dataNormalizer.New(logger)

	datumArray := []data.Datum{}
	for _, reference := range parser.References() {
		if datum := dataTypesFactory.ParseDatum(parser.WithReferenceObjectParser(reference)); datum != nil && *datum != nil {
			(*datum).Validate(validator.WithReference(strconv.Itoa(reference)))
			(*datum).Normalize(normalizer.WithReference(strconv.Itoa(reference)))
			datumArray = append(datumArray, *datum)
		}
	}

	if err = parser.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
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
		datum.SetProvenance(GetProvenanceFromRequest(ctx, req, authDetails))
	}

	if deduplicator, getErr := dataServiceContext.DataDeduplicatorFactory().Get(ctx, dataSet); getErr != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get deduplicator", getErr)
		return
	} else if deduplicator == nil {
		dataServiceContext.RespondWithInternalServerFailure("Deduplicator not found")
		return
	} else if err = deduplicator.AddData(ctx, dataSet, datumArray); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to add data", err)
		return
	}

	updatesSummary := make(map[string]struct{})
	for _, datum := range datumArray {
		summary.CheckDatumUpdatesSummary(updatesSummary, datum)
	}
	summary.MaybeUpdateSummary(ctx, dataServiceContext.SummarizerRegistry(), updatesSummary, *dataSet.UserID, summaryTypes.OutdatedReasonDataAdded)

	if err = dataServiceContext.MetricClient().RecordMetric(ctx, "data_sets_data_create", map[string]string{"count": strconv.Itoa(len(datumArray))}); err != nil {
		lgr.WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, []struct{}{})
}
