package service

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"net/http"
)

func ErrorInternalServerFailure() *Error {
	return &Error{
		Code:   "internal-server-failure",
		Status: http.StatusInternalServerError,
		Title:  "internal server failure",
		Detail: "Internal server failure",
	}
}
