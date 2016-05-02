package dataservices

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
	ErrorJSONMalformed      = "json-malformed"
	ErrorUserIDMalformed    = "user-id-malformed"
	ErrorUserIDNotFound     = "user-id-not-found"
	ErrorDatasetIDMalformed = "dataset-id-malformed"
	ErrorDatasetIDNotFound  = "dataset-id-not-found"
	ErrorDatasetClosed      = "dataset-closed"
)

var errorTemplates = map[string]*service.Error{
	ErrorJSONMalformed: {
		Status: http.StatusBadRequest,
		Title:  "json is malformed",
		Detail: "JSON is malformed",
	},
	ErrorUserIDMalformed: {
		Status: http.StatusBadRequest,
		Title:  "user id is malformed",
		Detail: "User id '%s' is malformed",
	},
	ErrorUserIDNotFound: {
		Status: http.StatusNotFound,
		Title:  "user with specified id not found",
		Detail: "User with id '%s' not found",
	},
	ErrorDatasetIDMalformed: {
		Status: http.StatusBadRequest,
		Title:  "dataset id is malformed",
		Detail: "Dataset id '%s' is malformed",
	},
	ErrorDatasetIDNotFound: {
		Status: http.StatusNotFound,
		Title:  "dataset with specified id not found",
		Detail: "Dataset with id '%s' not found",
	},
	ErrorDatasetClosed: {
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
