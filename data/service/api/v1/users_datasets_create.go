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

// UsersDataSetsCreate godoc
// @Summary Create a data sets
// @Description Create a new data sets.
// @Description Caller must be a service, the owner, or have the authorizations to do it in behalf of the user.
// @ID platform-data-api-UsersDataSetsCreate
// @Accept json
// @Produce json
// @Param userId path string true "user ID"
// @Param usersDataSetsCreateParams body data.DataSetCreate true "The new data set information"
// @Security TidepoolSessionToken
// @Security TidepoolServiceSecret
// @Security TidepoolAuthorization
// @Security TidepoolRestrictedToken
// @Success 200 {object} upload.Upload "Operation is a success"
// @Failure 400 {object} service.Error "User id is missing or JSON body is malformed"
// @Failure 403 {object} service.Error "Forbiden: caller is not authorized"
// @Failure 500 {object} service.Error "Unable to perform the operation"
// @Router /v1/users/:userId/datasets [post]
func UsersDataSetsCreate(dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	lgr := log.LoggerFromContext(ctx)

	targetUserID := dataServiceContext.Request().PathParam("userId")
	if targetUserID == "" {
		dataServiceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	if details := request.DetailsFromContext(ctx); !details.IsService() {
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

	parser := structureParser.NewObject(&rawDatum)
	validator := structureValidator.New()
	normalizer := dataNormalizer.New()

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

	dataSet.Normalize(normalizer)

	if err := normalizer.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
		return
	}

	dataSet.DataState = pointer.FromString("open") // TODO: Deprecated DataState (after data migration)
	dataSet.State = pointer.FromString("open")

	if err := dataServiceContext.DataSession().CreateDataSet(ctx, dataSet); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to insert data set", err)
		return
	}

	if deduplicator, err := dataServiceContext.DataDeduplicatorFactory().New(dataSet); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get deduplicator", err)
		return
	} else if deduplicator == nil {
		dataServiceContext.RespondWithInternalServerFailure("Deduplicator not found", err)
		return
	} else if dataSet, err = deduplicator.Open(ctx, dataServiceContext.DataSession(), dataSet); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to open", err)
		return
	}

	if err := dataServiceContext.MetricClient().RecordMetric(ctx, "users_data_sets_create"); err != nil {
		lgr.WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusCreated, dataSet)
}
