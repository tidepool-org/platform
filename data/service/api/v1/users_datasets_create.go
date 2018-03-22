package v1

import (
	"net/http"

	"github.com/tidepool-org/platform/data/context"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

func UsersDatasetsCreate(dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	lgr := log.LoggerFromContext(ctx)

	targetUserID := dataServiceContext.Request().PathParam("userId")
	if targetUserID == "" {
		dataServiceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	if details := request.DetailsFromContext(ctx); !details.IsService() {
		permissions, err := dataServiceContext.UserClient().GetUserPermissions(ctx, details.UserID(), targetUserID)
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

	var rawDatum map[string]interface{}
	if err := dataServiceContext.Request().DecodeJsonPayload(&rawDatum); err != nil {
		dataServiceContext.RespondWithError(service.ErrorJSONMalformed())
		return
	}

	datumContext, err := context.NewStandard(lgr)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create datum context", err)
		return
	}

	datumParser, err := parser.NewStandardObject(datumContext, dataServiceContext.DataFactory(), &rawDatum, parser.AppendErrorNotParsed)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create datum parser", err)
		return
	}

	validator := structureValidator.New()
	normalizer := dataNormalizer.New()

	datasetDatum, err := parser.ParseDatum(datumParser, dataServiceContext.DataFactory())
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to parse datum parser", err)
		return
	}

	if datasetDatum != nil && *datasetDatum != nil {
		datumParser.ProcessNotParsed()
		(*datasetDatum).Validate(validator)
	}

	if errs := datumContext.Errors(); len(errs) > 0 {
		dataServiceContext.RespondWithStatusAndErrors(http.StatusBadRequest, errs)
		return
	}

	if err = validator.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
		return
	}

	(*datasetDatum).SetUserID(&targetUserID)

	normalizer.Normalize(*datasetDatum)

	if err = normalizer.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
		return
	}

	dataset, ok := (*datasetDatum).(*upload.Upload)
	if !ok {
		dataServiceContext.RespondWithInternalServerFailure("Unexpected datum type", *datasetDatum)
		return
	}

	dataset.DataState = pointer.String("open") // TODO: Deprecated DataState (after data migration)
	dataset.ID = pointer.String(id.New())
	dataset.State = pointer.String("open")

	if err = dataServiceContext.DataSession().CreateDataset(ctx, dataset); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to insert dataset", err)
		return
	}

	deduplicator, err := dataServiceContext.DataDeduplicatorFactory().NewDeduplicatorForDataset(lgr, dataServiceContext.DataSession(), dataset)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create deduplicator for dataset", err)
		return
	}

	if err = deduplicator.RegisterDataset(ctx); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to register dataset with deduplicator", err)
		return
	}

	if err = dataServiceContext.MetricClient().RecordMetric(ctx, "users_datasets_create"); err != nil {
		lgr.WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusCreated, dataset)
}
