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

func ErrorDatasetIDMissing() *service.Error {
	return &service.Error{
		Code:   "dataset-id-missing",
		Status: http.StatusBadRequest,
		Title:  "dataset id is missing",
		Detail: "Dataset id is missing",
	}
}

func ErrorDatasetIDNotFound(datasetID string) *service.Error {
	return &service.Error{
		Code:   "dataset-id-not-found",
		Status: http.StatusNotFound,
		Title:  "dataset with specified id not found",
		Detail: fmt.Sprintf("Dataset with id %s not found", datasetID),
	}
}

func ErrorDatasetClosed(datasetID string) *service.Error {
	return &service.Error{
		Code:   "dataset-closed",
		Status: http.StatusConflict,
		Title:  "dataset with specified id is closed for new data",
		Detail: fmt.Sprintf("Dataset with id %s is closed for new data", datasetID),
	}
}
