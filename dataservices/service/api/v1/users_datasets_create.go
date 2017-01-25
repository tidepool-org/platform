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

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/dataservices/service"
	commonService "github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/userservices/client"
)

func UsersDatasetsCreate(serviceContext service.Context) {
	targetUserID := serviceContext.Request().PathParam("userid")
	if targetUserID == "" {
		serviceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	if !serviceContext.AuthenticationDetails().IsServer() {
		permissions, err := serviceContext.UserServicesClient().GetUserPermissions(serviceContext, serviceContext.AuthenticationDetails().UserID(), targetUserID)
		if err != nil {
			if client.IsUnauthorizedError(err) {
				serviceContext.RespondWithError(commonService.ErrorUnauthorized())
			} else {
				serviceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[client.UploadPermission]; !ok {
			serviceContext.RespondWithError(commonService.ErrorUnauthorized())
			return
		}
	}

	targetUserGroupID, err := serviceContext.UserServicesClient().GetUserGroupID(serviceContext, targetUserID)
	if err != nil {
		if client.IsUnauthorizedError(err) {
			serviceContext.RespondWithError(commonService.ErrorUnauthorized())
		} else {
			serviceContext.RespondWithInternalServerFailure("Unable to get group id for target user", err)
		}
		return
	}

	var rawDatum map[string]interface{}
	if err = serviceContext.Request().DecodeJsonPayload(&rawDatum); err != nil {
		serviceContext.RespondWithError(commonService.ErrorJSONMalformed())
		return
	}

	datumContext, err := context.NewStandard(serviceContext.Logger())
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to create datum context", err)
		return
	}

	datumParser, err := parser.NewStandardObject(datumContext, serviceContext.DataFactory(), &rawDatum, parser.AppendErrorNotParsed)
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to create datum parser", err)
		return
	}

	datumValidator, err := validator.NewStandard(datumContext)
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to create datum validator", err)
		return
	}

	datumNormalizer, err := normalizer.NewStandard(datumContext)
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to create datum normalizer", err)
		return
	}

	datasetDatum, err := parser.ParseDatum(datumParser, serviceContext.DataFactory())
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to parse datum parser", err)
		return
	}

	if datasetDatum != nil && *datasetDatum != nil {
		datumParser.ProcessNotParsed()
		(*datasetDatum).Validate(datumValidator)
	}

	if errors := datumContext.Errors(); len(errors) > 0 {
		serviceContext.RespondWithStatusAndErrors(http.StatusBadRequest, errors)
		return
	}

	(*datasetDatum).SetUserID(targetUserID)
	(*datasetDatum).SetGroupID(targetUserGroupID)
	(*datasetDatum).Normalize(datumNormalizer)

	dataset, ok := (*datasetDatum).(*upload.Upload)
	if !ok {
		serviceContext.RespondWithInternalServerFailure("Unexpected datum type", *datasetDatum)
		return
	}

	if err = serviceContext.DataStoreSession().CreateDataset(dataset); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to insert dataset", err)
		return
	}

	deduplicator, err := serviceContext.DataDeduplicatorFactory().NewDeduplicatorForDataset(serviceContext.Logger(), serviceContext.DataStoreSession(), dataset)
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to create deduplicator for dataset", err)
		return
	}

	if err = deduplicator.RegisterDataset(); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to register dataset with deduplicator", err)
		return
	}

	if err = serviceContext.MetricServicesClient().RecordMetric(serviceContext, "users_datasets_create"); err != nil {
		serviceContext.Logger().WithError(err).Error("Unable to record metric")
	}

	serviceContext.RespondWithStatusAndData(http.StatusCreated, dataset)
}
