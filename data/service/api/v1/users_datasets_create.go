package v1

import (
	"net/http"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func UsersDataSetsCreate(dataServiceContext dataService.Context) {
	req := dataServiceContext.Request()
	ctx := req.Context()
	lgr := log.LoggerFromContext(ctx)

	targetUserID := dataServiceContext.Request().PathParam("userId")
	if targetUserID == "" {
		dataServiceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	var details = request.GetAuthDetails(ctx)
	if !details.IsService() {
		permissions, err := dataServiceContext.PermissionClient().GetUserPermissions(ctx, details.UserID(), targetUserID)
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

	var rawDatum map[string]interface{}
	if err := dataServiceContext.Request().DecodeJsonPayload(&rawDatum); err != nil {
		dataServiceContext.RespondWithError(service.ErrorJSONMalformed())
		return
	}

	logger := log.LoggerFromContext(ctx)
	parser := structureParser.NewObject(logger, &rawDatum)
	validator := structureValidator.New(logger)
	normalizer := dataNormalizer.New(logger)

	dataSet := upload.ParseUpload(parser)
	if dataSet != nil {
		parser.NotParsed()
	}

	if err := parser.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
		return
	}

	dataSet.Validate(validator)
	if err := validator.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
		return
	}

	dataSet.SetUserID(&targetUserID)
	if details.IsUser() {
		dataSet.SetCreatedUserID(pointer.FromString(details.UserID()))
	}

	dataSet.Normalize(normalizer)

	if err := normalizer.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
		return
	}

	dataSet.DataState = pointer.FromString("open") // TODO: Deprecated DataState (after data migration)
	dataSet.State = pointer.FromString("open")
	dataSet.Provenance = CollectProvenanceInfo(ctx, req, details)

	if err := dataServiceContext.DataRepository().CreateDataSet(ctx, dataSet); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to insert data set", err)
		return
	}

	if deduplicator, err := dataServiceContext.DataDeduplicatorFactory().New(ctx, dataSet); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get deduplicator", err)
		return
	} else if deduplicator == nil {
		dataServiceContext.RespondWithInternalServerFailure("Deduplicator not found", err)
		return
	} else if dataSet, err = deduplicator.Open(ctx, dataServiceContext.DataRepository(), dataSet); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to open", err)
		return
	}

	if err := dataServiceContext.MetricClient().RecordMetric(ctx, "users_data_sets_create"); err != nil {
		lgr.WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusCreated, dataSet)
}
