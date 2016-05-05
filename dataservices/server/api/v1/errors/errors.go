package errors

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

const (
	AuthenticationTokenMissing = "authentication-token-missing"
	Unauthenticated            = "unauthenticated"
	Unauthorized               = "unauthorized"
	JSONMalformed              = "json-malformed"
	UserIDMissing              = "user-id-missing"
	UserIDNotFound             = "user-id-not-found"
	DatasetIDMissing           = "dataset-id-missing"
	DatasetIDNotFound          = "dataset-id-not-found"
	DatasetClosed              = "dataset-closed"
)

var errorTemplates = map[string]*service.Error{
	AuthenticationTokenMissing: {
		Status: http.StatusUnauthorized,
		Title:  "authentication token missing",
		Detail: "Authentication token missing",
	},
	Unauthenticated: {
		Status: http.StatusUnauthorized,
		Title:  "authentication token is invalid",
		Detail: "Authentication token is invalid",
	},
	Unauthorized: {
		Status: http.StatusForbidden,
		Title:  "authentication token is not authorized for requested action",
		Detail: "Authentication token is not authorized for requested action",
	},
	JSONMalformed: {
		Status: http.StatusBadRequest,
		Title:  "json is malformed",
		Detail: "JSON is malformed",
	},
	UserIDMissing: {
		Status: http.StatusBadRequest,
		Title:  "user id is missing",
		Detail: "User id is missing",
	},
	UserIDNotFound: {
		Status: http.StatusNotFound,
		Title:  "user with specified id not found",
		Detail: "User with id '%s' not found",
	},
	DatasetIDMissing: {
		Status: http.StatusBadRequest,
		Title:  "dataset id is missing",
		Detail: "Dataset id is missing",
	},
	DatasetIDNotFound: {
		Status: http.StatusNotFound,
		Title:  "dataset with specified id not found",
		Detail: "Dataset with id '%s' not found",
	},
	DatasetClosed: {
		Status: http.StatusConflict,
		Title:  "dataset with specified id is closed for new data",
		Detail: "Dataset with id '%s' is closed for new data",
	},
}

func ConstructError(code string, args ...interface{}) *service.Error {
	if err, ok := errorTemplates[code]; ok {
		err := err.Clone()
		err.Code = code
		if len(args) != 0 {
			err.Detail = fmt.Sprintf(err.Detail, args...)
		}
		return err
	}
	return nil
}

func ConstructParameterError(code string, parameter string, args ...interface{}) *service.Error {
	err := ConstructError(code, args)
	if err != nil {
		err.Source = &service.Source{Parameter: parameter}
	}
	return err
}

func ConstructPointerError(code string, pointer string, args ...interface{}) *service.Error {
	err := ConstructError(code, args)
	if err != nil {
		err.Source = &service.Source{Pointer: pointer}
	}
	return err
}
