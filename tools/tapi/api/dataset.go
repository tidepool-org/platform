package api

import (
	"errors"
	"net/http"
	"strconv"
)

type (
	Filter struct {
		Deleted *bool
	}

	Pagination struct {
		Page *int
		Size *int
	}
)

func (a *API) ListDataSets(userID string, filter *Filter, pagination *Pagination) (*ResponseArray, error) {
	if userID == "" {
		var err error
		userID, err = a.fetchSessionUserID()
		if err != nil {
			return nil, err
		}
	}

	queryMap := map[string]string{}
	if filter != nil {
		if filter.Deleted != nil {
			queryMap["deleted"] = strconv.FormatBool(*filter.Deleted)
		}
	}
	if pagination != nil {
		if pagination.Page != nil {
			queryMap["page"] = strconv.Itoa(*pagination.Page)
		}
		if pagination.Size != nil {
			queryMap["size"] = strconv.Itoa(*pagination.Size)
		}
	}

	return a.asResponseArray(a.request("GET", a.addQuery(a.joinPaths("v1", "users", userID, "data_sets"), queryMap),
		requestFuncs{a.addSessionToken()},
		responseFuncs{a.expectStatusCode(http.StatusOK)}))
}

func (a *API) DeleteDataSet(dataSetID string) error {
	if dataSetID == "" {
		return errors.New("Data set id is missing")
	}

	return a.asEmpty(a.request("DELETE", a.joinPaths("v1", "data_sets", dataSetID),
		requestFuncs{a.addSessionToken()},
		responseFuncs{a.expectStatusCode(http.StatusOK)}))
}
