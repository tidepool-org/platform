package request

import (
	"net/http"

	"github.com/tidepool-org/platform/errors"
)

const (
	ErrorCodeUnexpectedResponse = "unexpected-response"
	ErrorCodeBadRequest         = "bad-request"
	ErrorCodeUnauthenticated    = "unauthenticated"
	ErrorCodeUnauthorized       = "unauthorized"
	ErrorCodeResourceNotFound   = "resource-not-found"
	ErrorCodeParameterMissing   = "parameter-missing"
	ErrorCodeJSONMalformed      = "json-malformed"
)

func ErrorUnexpectedResponse(res *http.Response, req *http.Request) error {
	return errors.Preparedf(ErrorCodeUnexpectedResponse, "unexpected response", "unexpected response status code %d from %s %q", res.StatusCode, req.Method, req.URL.String())
}

func ErrorBadRequest() error {
	return errors.Prepared(ErrorCodeBadRequest, "bad request", "bad request")
}

func ErrorUnauthenticated() error {
	return errors.Prepared(ErrorCodeUnauthenticated, "authentication token is invalid", "authentication token is invalid")
}

func ErrorUnauthorized() error {
	return errors.Prepared(ErrorCodeUnauthorized, "authentication token is not authorized for requested action", "authentication token is not authorized for requested action")
}

func ErrorResourceNotFound() error {
	return errors.Prepared(ErrorCodeResourceNotFound, "resource not found", "resource not found")
}

func ErrorResourceNotFoundWithID(id string) error {
	return errors.Preparedf(ErrorCodeResourceNotFound, "resource not found", "resource with id %q not found", id)
}

func ErrorParameterMissing(parameter string) error {
	return errors.Preparedf(ErrorCodeParameterMissing, "parameter is missing", "parameter %q is missing", parameter)
}

func ErrorJSONMalformed() error {
	return errors.Prepared(ErrorCodeJSONMalformed, "json is malformed", "json is malformed")
}

func StatusCodeForError(err error) int {
	if err != nil {
		switch errors.Code(err) {
		case ErrorCodeBadRequest:
			return http.StatusBadRequest
		case ErrorCodeUnauthenticated:
			return http.StatusUnauthorized
		case ErrorCodeUnauthorized:
			return http.StatusForbidden
		case ErrorCodeResourceNotFound:
			return http.StatusNotFound
		}
	}
	return http.StatusInternalServerError
}
