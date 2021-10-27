package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/tidepool-org/platform/data"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataService "github.com/tidepool-org/platform/data/service"
	dataTypesFactory "github.com/tidepool-org/platform/data/types/factory"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func DataSetsDataUpdate(dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()

	dataID := dataServiceContext.Request().PathParam("dataId")
	if dataID == "" {
		dataServiceContext.RespondWithError(ErrorDataIDMissing())
		return
	}
	dataSetID := dataServiceContext.Request().PathParam("dataSetId")
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

	if details := request.DetailsFromContext(ctx); !details.IsService() {
		var permissions permission.Permissions
		permissions, err := dataServiceContext.PermissionClient().GetUserPermissions(ctx, details.UserID(), *dataSet.UserID)
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

	rawDatum, err := dataServiceContext.DataRepository().GetDataSetDataByID(ctx, dataSetID, dataID)
	if err != nil {
		dataServiceContext.RespondWithError(ErrorDataSetDataMissing())
		return
	}

	// the result coming from the db decoder cannot be parsed properly a few lines down
	// but the json marshaled/unmarshaled can so doing this extra step for now
	resultsJSON, err := json.Marshal([]interface{}{rawDatum})
	if err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
		return
	}
	var rawDatumArrayDB []interface{}
	err = json.Unmarshal(resultsJSON, &rawDatumArrayDB)
	if err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
		return
	}

	parser := structureParser.NewArray(&rawDatumArrayDB)
	validator := structureValidator.New()
	normalizer := dataNormalizer.New()

	datumArrayDB := []data.Datum{}
	for _, reference := range parser.References() {
		if datum := dataTypesFactory.ParseDatum(parser.WithReferenceObjectParser(reference)); datum != nil && *datum != nil {
			(*datum).Validate(validator.WithReference(strconv.Itoa(reference)))
			(*datum).Normalize(normalizer.WithReference(strconv.Itoa(reference)))
			datumArrayDB = append(datumArrayDB, *datum)
		}
	}
	parser.NotParsed()

	if err = parser.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
		return
	}

	var rawDatumWeb interface{}
	if err = dataServiceContext.Request().DecodeJsonPayload(&rawDatumWeb); err != nil {
		dataServiceContext.RespondWithError(service.ErrorJSONMalformed())
		return
	}
	rawDatumArray := []interface{}{rawDatumWeb}

	parser = structureParser.NewArray(&rawDatumArray)
	validator = structureValidator.New()
	normalizer = dataNormalizer.New()

	datumArray := []data.Datum{}
	for _, reference := range parser.References() {
		if datum := dataTypesFactory.ParseDatum(parser.WithReferenceObjectParser(reference)); datum != nil && *datum != nil {
			(*datum).Validate(validator.WithReference(strconv.Itoa(reference)))
			(*datum).Normalize(normalizer.WithReference(strconv.Itoa(reference)))
			datumArray = append(datumArray, *datum)
		}
	}
	parser.NotParsed()

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
	}

	dbDatum := datumArrayDB[0]
	datum := datumArray[0]

	// incoming datum replaces the db datum, while the previous dbDatum
	// and its history get flattened into a history array and added to the new entry
	existingHistory := *dbDatum.GetHistory()
	dbDatum.SetHistory(nil)
	existingHistory = append(existingHistory, dbDatum)
	datum.SetHistory(&existingHistory)
	datum.SetID(&dataID)

	err = dataServiceContext.DataRepository().UpdateDataSetData(ctx, dataSet, datum)
	if err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
		return
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, struct{}{})
}
