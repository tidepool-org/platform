package v1

import (
	"net/http"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
	userClient "github.com/tidepool-org/platform/user/client"
)

func UsersDatasetsCreate(dataServiceContext dataService.Context) {
	targetUserID := dataServiceContext.Request().PathParam("userid")
	if targetUserID == "" {
		dataServiceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	if !dataServiceContext.AuthDetails().IsServer() {
		permissions, err := dataServiceContext.UserClient().GetUserPermissions(dataServiceContext, dataServiceContext.AuthDetails().UserID(), targetUserID)
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

	var rawDatum map[string]interface{}
	if err := dataServiceContext.Request().DecodeJsonPayload(&rawDatum); err != nil {
		dataServiceContext.RespondWithError(service.ErrorJSONMalformed())
		return
	}

	datumContext, err := context.NewStandard(dataServiceContext.Logger())
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create datum context", err)
		return
	}

	datumParser, err := parser.NewStandardObject(datumContext, dataServiceContext.DataFactory(), &rawDatum, parser.AppendErrorNotParsed)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create datum parser", err)
		return
	}

	datumValidator, err := validator.NewStandard(datumContext)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create datum validator", err)
		return
	}

	datumNormalizer, err := normalizer.NewStandard(datumContext)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create datum normalizer", err)
		return
	}

	datasetDatum, err := parser.ParseDatum(datumParser, dataServiceContext.DataFactory())
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to parse datum parser", err)
		return
	}

	if datasetDatum != nil && *datasetDatum != nil {
		datumParser.ProcessNotParsed()
		(*datasetDatum).Validate(datumValidator)
	}

	if errors := datumContext.Errors(); len(errors) > 0 {
		dataServiceContext.RespondWithStatusAndErrors(http.StatusBadRequest, errors)
		return
	}

	(*datasetDatum).SetUserID(targetUserID)
	(*datasetDatum).Normalize(datumNormalizer)

	dataset, ok := (*datasetDatum).(*upload.Upload)
	if !ok {
		dataServiceContext.RespondWithInternalServerFailure("Unexpected datum type", *datasetDatum)
		return
	}

	if err = dataServiceContext.DataStoreSession().CreateDataset(dataset); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to insert dataset", err)
		return
	}

	deduplicator, err := dataServiceContext.DataDeduplicatorFactory().NewDeduplicatorForDataset(dataServiceContext.Logger(), dataServiceContext.DataStoreSession(), dataset)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create deduplicator for dataset", err)
		return
	}

	if err = deduplicator.RegisterDataset(); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to register dataset with deduplicator", err)
		return
	}

	if err = dataServiceContext.MetricClient().RecordMetric(dataServiceContext, "users_datasets_create"); err != nil {
		dataServiceContext.Logger().WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusCreated, dataset)
}
