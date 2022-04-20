package v1

import (
	"net/http"
	"strconv"

	"github.com/tidepool-org/platform/data"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataService "github.com/tidepool-org/platform/data/service"
	dataTypesFactory "github.com/tidepool-org/platform/data/types/factory"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type EmptyBody struct{}

// DataSetsDataCreate godoc
// @Summary Add data to a DataSets
// @ID platform-data-api-DataSetsDataCreate
// @Accept json
// @Produce json
// @Param dataSetID path string true "dataSet ID"
// @Param data body []types.Base true "Array of data, of one type only"
// @Security TidepoolSessionToken
// @Security TidepoolServiceSecret
// @Security TidepoolAuthorization
// @Security TidepoolRestrictedToken
// @Success 200 {object} EmptyBody "Operation is a success"
// @Failure 400 {object} service.Error "Data set id is missing"
// @Failure 403 {object} service.Error "Auth token is not authorized for requested action"
// @Failure 404 {object} service.Error "Data set with specified id not found"
// @Failure 409 {object} service.Error "Data set with specified id is closed"
// @Failure 500 {object} service.Error "Unable to perform the operation"
// @Router /v1/datasets/:dataSetId/data [post]
func DataSetsDataCreate(dataServiceContext dataService.Context) {
	req := dataServiceContext.Request()
	ctx := req.Context()

	dataSetID := req.PathParam("dataSetId")
	if dataSetID == "" {
		dataServiceContext.RespondWithError(ErrorDataSetIDMissing())
		return
	}

	dataSet, err := dataServiceContext.DataRepository().GetDataSetByID(ctx, dataSetID)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get data set by id", err)
		return
	}
	if dataSet == nil {
		dataServiceContext.RespondWithError(ErrorDataSetIDNotFound(dataSetID))
		return
	}

	permissions, err := dataServiceContext.PermissionClient().GetUserPermissions(req, *dataSet.UserID)
	if err != nil {
		if request.IsErrorUnauthorized(err) {
			dataServiceContext.RespondWithError(service.ErrorUnauthorized())
		} else {
			dataServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
		}
		return
	}
	if !permissions {
		dataServiceContext.RespondWithError(service.ErrorUnauthorized())
		return
	}

	if (dataSet.State != nil && *dataSet.State == "closed") || (dataSet.DataState != nil && *dataSet.DataState == "closed") { // TODO: Deprecated DataState (after data migration)
		dataServiceContext.RespondWithError(ErrorDataSetClosed(dataSetID))
		return
	}

	var rawDatumArray []interface{}
	if err = req.DecodeJsonPayload(&rawDatumArray); err != nil {
		dataServiceContext.RespondWithError(service.ErrorJSONMalformed())
		return
	}

	parser := structureParser.NewArray(&rawDatumArray)
	validator := structureValidator.New()
	normalizer := dataNormalizer.New()

	datumArray := []data.Datum{}
	for _, reference := range parser.References() {
		if datum := dataTypesFactory.ParseDatum(parser.WithReferenceObjectParser(reference)); datum != nil && *datum != nil {
			(*datum).Validate(validator.WithReference(strconv.Itoa(reference)))
			if (*datum).IsValid(validator.WithReference(strconv.Itoa(reference))) {
				(*datum).Normalize(normalizer.WithReference(strconv.Itoa(reference)))
				datumArray = append(datumArray, *datum)
			} else {
				// reset Warning
				validator.ResetWarning()
			}
		}
	}
	parser.NotParsed()

	if err = parser.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), req).Error(http.StatusBadRequest, err)
		return
	}
	if err = validator.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), req).Error(http.StatusBadRequest, err)
		return
	}
	if err = normalizer.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), req).Error(http.StatusBadRequest, err)
		return
	}

	datumArray = append(datumArray, normalizer.Data()...)

	for _, datum := range datumArray {
		datum.SetUserID(dataSet.UserID)
		datum.SetDataSetID(dataSet.UploadID)
	}

	if deduplicator, getErr := dataServiceContext.DataDeduplicatorFactory().Get(dataSet); getErr != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get deduplicator", getErr)
		return
	} else if deduplicator == nil {
		dataServiceContext.RespondWithInternalServerFailure("Deduplicator not found")
		return
	} else if err = deduplicator.AddData(ctx, dataServiceContext.DataRepository(), dataSet, datumArray); err != nil { // write in DB
		dataServiceContext.RespondWithInternalServerFailure("Unable to add data", err)
		return
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, []struct{}{})
}
