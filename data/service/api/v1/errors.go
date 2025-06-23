package v1

import (
	"fmt"
	"net/http"

	"github.com/tidepool-org/platform/service"
)

func ErrorUserIDMissing() *service.Error {
	return &service.Error{
		Code:   "user-id-missing",
		Status: http.StatusBadRequest,
		Title:  "user id is missing",
		Detail: "User id is missing",
	}
}

func ErrorUserIDNotFound(userID string) *service.Error {
	return &service.Error{
		Code:   "user-id-not-found",
		Status: http.StatusNotFound,
		Title:  "user with specified id not found",
		Detail: fmt.Sprintf("User with id %s not found", userID),
	}
}

func ErrorDataSetIDMissing() *service.Error {
	return &service.Error{
		Code:   "data-set-id-missing",
		Status: http.StatusBadRequest,
		Title:  "data set id is missing",
		Detail: "Data set id is missing",
	}
}

func ErrorDataSetIDNotFound(dataSetID string) *service.Error {
	return &service.Error{
		Code:   "data-set-id-not-found",
		Status: http.StatusNotFound,
		Title:  "data set with specified id not found",
		Detail: fmt.Sprintf("Data set with id %s not found", dataSetID),
	}
}

func ErrorDataSetClosed(dataSetID string) *service.Error {
	return &service.Error{
		Code:   "data-set-closed",
		Status: http.StatusConflict,
		Title:  "data set with specified id is closed for new data",
		Detail: fmt.Sprintf("Data set with id %s is closed for new data", dataSetID),
	}
}

func ErrorTidepoolLinkIDMissing() *service.Error {
	return &service.Error{
		Code:   "tidepool-link-id-missing",
		Status: http.StatusBadRequest,
		Title:  "tidepool link id is missing",
		Detail: "Tidepool link id is missing",
	}
}

func ErrorTidepoolLinkIDNotFound() *service.Error {
	return &service.Error{
		Code:   "tidepool-link-id-not-found",
		Status: http.StatusNotFound,
		Title:  "tidepool link id not found",
		Detail: "Tidepool link id not found",
	}
}
