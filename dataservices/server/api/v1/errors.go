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
	"fmt"
	"net/http"

	"github.com/tidepool-org/platform/service"
)

func ErrorAuthenticationTokenMissing() *service.Error {
	return &service.Error{
		Code:   "authentication-token-missing",
		Status: http.StatusUnauthorized,
		Title:  "authentication token missing",
		Detail: "Authentication token missing",
	}
}

func ErrorUnauthenticated() *service.Error {
	return &service.Error{
		Code:   "unauthenticated",
		Status: http.StatusUnauthorized,
		Title:  "authentication token is invalid",
		Detail: "Authentication token is invalid",
	}
}

func ErrorUnauthorized() *service.Error {
	return &service.Error{
		Code:   "unauthorized",
		Status: http.StatusForbidden,
		Title:  "authentication token is not authorized for requested action",
		Detail: "Authentication token is not authorized for requested action",
	}
}

func ErrorJSONMalformed() *service.Error {
	return &service.Error{
		Code:   "json-malformed",
		Status: http.StatusBadRequest,
		Title:  "json is malformed",
		Detail: "JSON is malformed",
	}
}

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
