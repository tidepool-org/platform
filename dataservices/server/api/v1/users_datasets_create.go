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
	"github.com/tidepool-org/platform/data/deduplicator/root"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/dataservices/server"
	"github.com/tidepool-org/platform/userservices/client"
)

func UsersDatasetsCreate(serverContext server.Context) {
	targetUserID := serverContext.Request().PathParam("userid")
	if targetUserID == "" {
		serverContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	err := serverContext.Client().ValidateTargetUserPermissions(serverContext, serverContext.RequestUserID(), targetUserID, client.UploadPermissions)
	if err != nil {
		if client.IsUnauthorizedError(err) {
			serverContext.RespondWithError(ErrorUnauthorized())
		} else {
			serverContext.RespondWithInternalServerFailure("Unable to validate target user permissions", err)
		}
		return
	}

	targetUserGroupID, err := serverContext.Client().GetUserGroupID(serverContext, targetUserID)
	if err != nil {
		if client.IsUnauthorizedError(err) {
			serverContext.RespondWithError(ErrorUnauthorized())
		} else {
			serverContext.RespondWithInternalServerFailure("Unable to get group id for target user", err)
		}
		return
	}

	var rawDatum map[string]interface{}
	if err = serverContext.Request().DecodeJsonPayload(&rawDatum); err != nil {
		serverContext.RespondWithError(ErrorJSONMalformed())
		return
	}

	datumContext, err := context.NewStandard(serverContext.Logger())
	if err != nil {
		serverContext.RespondWithInternalServerFailure("Unable to create datum context", err)
		return
	}

	datumParser, err := parser.NewStandardObject(datumContext, &rawDatum, parser.AppendErrorNotParsed)
	if err != nil {
		serverContext.RespondWithInternalServerFailure("Unable to create datum parser", err)
		return
	}

	datumValidator, err := validator.NewStandard(datumContext)
	if err != nil {
		serverContext.RespondWithInternalServerFailure("Unable to create datum validator", err)
		return
	}

	datumNormalizer, err := normalizer.NewStandard(datumContext)
	if err != nil {
		serverContext.RespondWithInternalServerFailure("Unable to create datum normalizer", err)
		return
	}

	datasetDatum, err := types.Parse(datumParser)
	if err != nil {
		serverContext.RespondWithInternalServerFailure("Unable to parse datum parser", err)
		return
	}

	datumParser.ProcessNotParsed()

	datasetDatum.Validate(datumValidator)

	if errors := datumContext.Errors(); len(errors) > 0 {
		serverContext.RespondWithStatusAndErrors(http.StatusBadRequest, errors)
		return
	}

	datasetDatum.SetUserID(targetUserID)
	datasetDatum.SetGroupID(targetUserGroupID)
	datasetDatum.Normalize(datumNormalizer)

	datasetUpload, ok := datasetDatum.(*upload.Upload)
	if !ok {
		serverContext.RespondWithInternalServerFailure("Unexpected datum type", datasetDatum)
		return
	}

	if err = serverContext.Store().Insert(datasetUpload); err != nil {
		serverContext.RespondWithInternalServerFailure("Unable to insert dataset", err)
		return
	}

	deduplicator, err := root.NewFactory().NewDeduplicator(serverContext.Logger(), serverContext.Store(), datasetUpload)
	if err != nil {
		serverContext.RespondWithInternalServerFailure("No duplicator found matching dataset", err)
		return
	}

	if err = deduplicator.InitializeDataset(); err != nil {
		serverContext.RespondWithInternalServerFailure("Unable to initialize dataset", err)
		return
	}

	// TODO: Filter datasetUpload to only "public" fields
	serverContext.Response().WriteHeader(http.StatusCreated)
	serverContext.Response().WriteJson(datasetUpload)
}
