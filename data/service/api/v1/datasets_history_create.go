package v1

import (
	"encoding/json"
	"fmt"
	system_log "log"
	"net/http"
	"strconv"

	"github.com/tidepool-org/platform/data"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataService "github.com/tidepool-org/platform/data/service"
	dataTypesFactory "github.com/tidepool-org/platform/data/types/factory"
	"github.com/tidepool-org/platform/history"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func DataSetsHistoryCreate(dataServiceContext dataService.Context) {
	system_log.Println("Inside creating history")
	// Get Ctx
	ctx := dataServiceContext.Request().Context()
	lgr := log.LoggerFromContext(ctx)

	// Get Parameters - DataID
	dataID := dataServiceContext.Request().PathParam("dataId")
	if dataID == "" {
		dataServiceContext.RespondWithError(ErrorDataIDMissing())
		return
	}

	// Get the Data from input data ID from the database
	// input data will be in form of []interface{}
	results, err := dataServiceContext.DataSession().GetDataByID(ctx, dataID)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get data by id for user", err)
		return
	}

	// We have to take the data to json and then back to a object structure.  Otherwise - we will have problems in
	// the parsing step
	system_log.Println("Results: ", results)
	resultsJSON, err := json.Marshal(results)
	system_log.Println("Results Json: ", string(resultsJSON))
	if err != nil {
		dataServiceContext.RespondWithError(service.ErrorJSONMalformed())
		return
	}
	var rawResults map[string]interface{}
	err = json.Unmarshal(resultsJSON, &rawResults)
	if err != nil {
		fmt.Println("error unmarshalling json:", err)
		return
	}

	// We need to get the dataSetID - it is the uploadID for the data item
	dataSetID := rawResults["uploadId"].(string)
	system_log.Println("Upload ID: ", dataSetID)

	// Get the DataSet
	dataSet, err := dataServiceContext.DataSession().GetDataSetByID(ctx, dataSetID)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get data set by id", err)
		return
	}
	if dataSet == nil {
		dataServiceContext.RespondWithError(ErrorDataSetIDNotFound(dataSetID))
		return
	}

	// Verify dataSet is in good state
	if (dataSet.State != nil && *dataSet.State == "closed") || (dataSet.DataState != nil && *dataSet.DataState == "closed") { // TODO: Deprecated DataState (after data migration)
		dataServiceContext.RespondWithError(ErrorDataSetClosed(dataSetID))
		return
	}

	// Authenticate
	if details := request.DetailsFromContext(ctx); !details.IsService() {
		var permissions permission.Permissions
		permissions, err = dataServiceContext.PermissionClient().GetUserPermissions(ctx, details.UserID(), *dataSet.UserID)
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

	// Parse data into Datum structure.  We need to do this - because we are going to
	// create an object from this and history information and write it to database
	resultsArray := make([]interface{}, 1)
	resultsArray[0] = rawResults
	parser := structureParser.NewArray(&resultsArray)
	validator := structureValidator.New()
	normalizer := dataNormalizer.New()

	datumArray := []data.Datum{}
	for _, reference := range parser.References() {
		if datum := dataTypesFactory.ParseDatum(parser.WithReferenceObjectParser(reference)); datum != nil && *datum != nil {
			(*datum).Validate(validator.WithReference(strconv.Itoa(reference)))
			(*datum).Normalize(normalizer.WithReference(strconv.Itoa(reference)))
			datumArray = append(datumArray, *datum)
		}
	}
	parser.NotParsed()

	// Validate, normalize and check errors
	if err = parser.Error(); err != nil {
		system_log.Println("parser error: ", err)
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
		return
	}

	// Get the Json from the request
	var rawDatumArray []interface{}
	if err = dataServiceContext.Request().DecodeJsonPayload(&rawDatumArray); err != nil {
		dataServiceContext.RespondWithError(service.ErrorJSONMalformed())
		return
	}

	// Parse the json from the input request - should be a history object
	historyParser := structureParser.NewArray(&rawDatumArray)
	historyValidator := structureValidator.New()
	historyNormalizer := dataNormalizer.New()

	historyArray := []history.History{}
	for _, reference := range historyParser.References() {
		if historyDatum := history.ParseHistory(historyParser.WithReferenceObjectParser(reference)); historyDatum != nil {
			historyDatum.Validate(historyValidator.WithReference(strconv.Itoa(reference)), resultsJSON)
			historyArray = append(historyArray, *historyDatum)
		}
	}
	historyParser.NotParsed()
	system_log.Println("raw datum")
	system_log.Println(rawDatumArray)
	system_log.Println("finished parsing")
	system_log.Println("history array: ", historyArray)
	system_log.Println("error: ", historyParser.Error())

	// Validate, normalize and check errors
	if err = historyParser.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
		return
	}
	if err = historyValidator.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
		return
	}
	if err = historyNormalizer.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
		return
	}

	system_log.Println("datum")
	system_log.Println("dataID: ", dataID)

	// Set up the new object that we make out of the data we received from the database and the history object
	datumArray[0].SetHistory(&historyArray[0])

	system_log.Println("data for id")
	system_log.Println("datumArray", datumArray)

	// Write our newly created object to database
	if deduplicator, getErr := dataServiceContext.DataDeduplicatorFactory().Get(dataSet); getErr != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get deduplicator", getErr)
		return
	} else if deduplicator == nil {
		dataServiceContext.RespondWithInternalServerFailure("Deduplicator not found")
		return
	} else if err = deduplicator.AddData(ctx, dataServiceContext.DataSession(), dataSet, datumArray); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to add data", err)
		return
	}

	// Record metrics
	if err = dataServiceContext.MetricClient().RecordMetric(ctx, "data_sets_history_create", map[string]string{"count": strconv.Itoa(len(historyArray))}); err != nil {
		lgr.WithError(err).Error("Unable to record metric")
	}

	// Respond with data (NOTE: change back to return blank data when ready)
	dataServiceContext.RespondWithStatusAndData(http.StatusOK, datumArray)
	// dataServiceContext.RespondWithStatusAndData(http.StatusOK, []struct{}{})
}
