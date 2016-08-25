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
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/dataservices/service"
	"github.com/tidepool-org/platform/userservices/client"
)

func UsersDatasetsCreate(serviceContext service.Context) {
	targetUserID := serviceContext.Request().PathParam("userid")
	if targetUserID == "" {
		serviceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	permissions, err := serviceContext.UserServicesClient().GetUserPermissions(serviceContext, serviceContext.RequestUserID(), targetUserID)
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

	targetUserGroupID, err := serviceContext.UserServicesClient().GetUserGroupID(serviceContext, targetUserID)
	if err != nil {
		if client.IsUnauthorizedError(err) {
			serviceContext.RespondWithError(ErrorUnauthorized())
		} else {
			serviceContext.RespondWithInternalServerFailure("Unable to get group id for target user", err)
		}
		return
	}

	var rawDatum map[string]interface{}
	if err = serviceContext.Request().DecodeJsonPayload(&rawDatum); err != nil {
		serviceContext.RespondWithError(ErrorJSONMalformed())
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

	dataset.SetUploadUserID(serviceContext.RequestUserID())

	if err = serviceContext.DataStoreSession().CreateDataset(dataset); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to insert dataset", err)
		return
	}

	deduplicator, err := serviceContext.DataDeduplicatorFactory().NewDeduplicator(serviceContext.Logger(), serviceContext.DataStoreSession(), dataset)
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("No duplicator found matching dataset", err)
		return
	}

	if err = deduplicator.InitializeDataset(); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to initialize dataset", err)
		return
	}

	serviceContext.RespondWithStatusAndData(http.StatusCreated, dataset)
}
